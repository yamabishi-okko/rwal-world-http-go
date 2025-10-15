package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

type Session struct {
	SID       string
	Data      map[string]any
	ExpiresAt time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

var (
	db        *sql.DB
	port      = "8080"
	sessionTTL = 14 * 24 * time.Hour // デフォルト14日
)

// ---- ユーティリティ ----
func mustGetenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func mustAtoiEnv(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func newSID() (string, error) {
	b := make([]byte, 16) // 16 bytes -> 32 hex
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func now() time.Time { return time.Now().UTC() }

// ---- DB アクセス ----
func upsertSession(ctx context.Context, s *Session) error {
	j, err := json.Marshal(s.Data)
	if err != nil {
		return err
	}
	// INSERT ... ON DUPLICATE KEY UPDATE
	_, err = db.ExecContext(ctx, `
		INSERT INTO sessions (sid, data, expires_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
		  data       = VALUES(data),
		  expires_at = VALUES(expires_at),
		  updated_at = VALUES(updated_at)
	`, s.SID, string(j), s.ExpiresAt, s.CreatedAt, s.UpdatedAt)
	return err
}

func getSession(ctx context.Context, sid string) (*Session, error) {
	row := db.QueryRowContext(ctx, `
		SELECT sid, data, expires_at, created_at, updated_at
		FROM sessions
		WHERE sid = ? AND expires_at > ?
	`, sid, now())
	var (
		_sid, _data string
		_exp, _cr, _up time.Time
	)
	if err := row.Scan(&_sid, &_data, &_exp, &_cr, &_up); err != nil {
		return nil, err
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(_data), &m); err != nil {
		return nil, err
	}
	return &Session{SID: _sid, Data: m, ExpiresAt: _exp, CreatedAt: _cr, UpdatedAt: _up}, nil
}

func touchSession(ctx context.Context, sid string, extend time.Duration) error {
	_, err := db.ExecContext(ctx, `
		UPDATE sessions
		SET expires_at = ?, updated_at = ?
		WHERE sid = ?
	`, now().Add(extend), now(), sid)
	return err
}

// ---- セッション管理（Cookie: sid） ----
func getOrCreateSID(w http.ResponseWriter, r *http.Request) (string, *Session, error) {
	// Cookie 取得
	if c, err := r.Cookie("sid"); err == nil {
		// DB存在チェック
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if s, err := getSession(ctx, c.Value); err == nil {
			return c.Value, s, nil
		}
		// 期限切れ/存在なし -> 新規発行にフォールバック
	}

	// 新規発行
	sid, err := newSID()
	if err != nil {
		return "", nil, err
	}
	s := &Session{
		SID:       sid,
		Data:      map[string]any{"counter": 1},
		ExpiresAt: now().Add(sessionTTL),
		CreatedAt: now(),
		UpdatedAt: now(),
	}
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := upsertSession(ctx, s); err != nil {
		return "", nil, err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "sid",
		Value:    sid,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		// Secure: true, // 本番はHTTPSで有効に
	})
	return sid, s, nil
}

// ---- ハンドラ ----

func handleSet(w http.ResponseWriter, r *http.Request) {
	// sid を発行 or 取得。新規なら counter=1 で保存済み
	sid, s, err := getOrCreateSID(w, r)
	if err != nil {
		http.Error(w, "session create failed", http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]any{
		"sid": sid, "session": s.Data, "expires_at": s.ExpiresAt,
	})
}

func handleMe(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("sid")
	if err != nil {
		http.Error(w, "no session", http.StatusUnauthorized)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	s, err := getSession(ctx, c.Value)
	if err != nil {
		http.Error(w, "session expired or not found", http.StatusUnauthorized)
		return
	}
	// カウンタ増加（学習用の動き）
	if n, ok := s.Data["counter"].(float64); ok {
		s.Data["counter"] = int(n) + 1
	} else if n, ok := s.Data["counter"].(int); ok {
		s.Data["counter"] = n + 1
	} else {
		s.Data["counter"] = 1
	}
	s.ExpiresAt = now().Add(sessionTTL)
	s.UpdatedAt = now()
	if err := upsertSession(ctx, s); err != nil {
		http.Error(w, "session update failed", http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]any{
		"sid": s.SID, "session": s.Data, "expires_at": s.ExpiresAt,
	})
}

func handleEcho(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "only POST", http.StatusMethodNotAllowed)
		return
	}
	var v any
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	writeJSON(w, map[string]any{"you_sent": v})
}

// ---- ルータ/ミドルウェア ----

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t0 := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(t0))
	})
}

func router() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/set", handleSet) // sid発行 or 取得（新規ならDB作成）
	mux.HandleFunc("/me", handleMe)   // CookieのsidでDB参照→counter++
	mux.HandleFunc("/echo", handleEcho)
	return logging(mux)
}

func main() {
	// .env（環境変数）読み込み
	if p := mustGetenv("PORT", ""); p != "" {
		port = p
	}
	if hrs := mustAtoiEnv("SESSION_TTL_HOURS", 0); hrs > 0 {
		sessionTTL = time.Duration(hrs) * time.Hour
	}

	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		log.Fatal("MYSQL_DSN is required (e.g. user:pass@tcp(127.0.0.1:3306)/ch13_session_db?parseTime=true)")
	}

	// DB 接続
	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	if err := db.Ping(); err != nil {
		log.Fatal("db ping:", err)
	}

	addr := ":" + port
	fmt.Println("Start server at", addr)
	if err := http.ListenAndServe(addr, router()); err != nil {
		log.Fatal(err)
	}
}
