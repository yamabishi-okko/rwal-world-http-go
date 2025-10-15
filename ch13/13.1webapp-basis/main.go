package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ====== 超ミニ・セッション（学習用） ======
// 本番では Redis や DB を使う。ここでは Map + Mutex に保存するだけ。
type sessionData struct {
	Counter int       `json:"counter"`
	Created time.Time `json:"created"`
}

var (
	sessions   = map[string]*sessionData{}
	sessionsMu sync.RWMutex
)

func newSID() string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 24)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// ====== ミドルウェア（ログ・リカバー） ======

type middleware func(http.Handler) http.Handler

func chain(h http.Handler, m ...middleware) http.Handler {
	for i := len(m) - 1; i >= 0; i-- {
		h = m[i](h)
	}
	return h
}

func logging() middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
		})
	}
}

func recoverer() middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					http.Error(w, "internal server error", http.StatusInternalServerError)
					log.Printf("panic: %v", rec)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// ====== ハンドラ群 ======

func handleRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte("hello word"))
}

func handleJSON(w http.ResponseWriter, r *http.Request) {
	type resp struct {
		Message string `json:"message"`
		Time    string `json:"time"`
	}
	writeJSON(w, resp{Message: "Hi Hi HiHiHi Aru Aru Tankentai!!", Time: time.Now().UTC().Format(time.RFC3339)})
}

func handleEcho(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "only POST", http.StatusMethodNotAllowed)
		return
	}
	// JSON をそのまま受けて返す
	var v any
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	writeJSON(w, map[string]any{"you_sent": v})
}

func handleSetCookie(w http.ResponseWriter, r *http.Request) {
	sid := newSID()
	sessionsMu.Lock()
	sessions[sid] = &sessionData{Counter: 1, Created: time.Now()}
	sessionsMu.Unlock()

	http.SetCookie(w, &http.Cookie{
		Name:     "sid",
		Value:    sid,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		// Secure: true, // 本番は https 前提で有効にする
	})
	writeJSON(w, map[string]any{"session": sessions[sid]})
}

func handleWhoAmI(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("sid")
	if err != nil {
		http.Error(w, "no session", http.StatusUnauthorized)
		return
	}
	sessionsMu.Lock()
	defer sessionsMu.Unlock()
	s, ok := sessions[c.Value]
	if !ok {
		http.Error(w, "no session", http.StatusUnauthorized)
		return
	}
	s.Counter++
	writeJSON(w, map[string]any{"session": s})
}

// クエリ・パスパラメータの例: /search?q=go  と  /user/123
func handleSearch(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	writeJSON(w, map[string]string{"q": q})
}

func handleUser(w http.ResponseWriter, r *http.Request) {
	// パス: /user/{id}
	id := strings.TrimPrefix(r.URL.Path, "/user/")
	if id == "" || strings.Contains(id, "/") {
		http.NotFound(w, r)
		return
	}
	if _, err := strconv.Atoi(id); err != nil {
		http.Error(w, "id must be number", http.StatusBadRequest)
		return
	}
	writeJSON(w, map[string]string{"id": id})
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

// ====== ルータ（ディスパッチャ） ======

func router() http.Handler {
	mux := http.NewServeMux()

	// 静的ファイル: / で index.html を見せたいので FileServer を /public にアタッチ
	publicDir := filepath.Join(".", "public")
	fs := http.FileServer(http.Dir(publicDir))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	mux.HandleFunc("/", handleRoot)
	mux.HandleFunc("/json", handleJSON)
	mux.HandleFunc("/echo", handleEcho)
	mux.HandleFunc("/set-cookie", handleSetCookie)
	mux.HandleFunc("/whoami", handleWhoAmI)
	mux.HandleFunc("/search", handleSearch)
	mux.HandleFunc("/user/", handleUser)

	return chain(mux, logging(), recoverer())
}

func main() {
	rand.Seed(time.Now().UnixNano())
	addr := ":8080"
	fmt.Println("Start server at", addr)
	if err := http.ListenAndServe(addr, router()); err != nil {
		log.Fatal(err)
	}
}
