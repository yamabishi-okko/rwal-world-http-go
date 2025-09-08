package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"path/filepath"
	"time"
)

func main() {
	mux := http.NewServeMux()

	// 静的
	mux.Handle("/",
		http.FileServer(http.Dir("./public")))

	// JSON API（GET）
	mux.HandleFunc("/api/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(w).Encode(map[string]any{
			"message": "hello from fetch api",
			"time":    time.Now().Format(time.RFC3339),
		})
	})

	// ダウンロード（GET + Content-Disposition）
	mux.HandleFunc("/api/download", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Content-Disposition", "attachment; filename=\"sample.txt\"")
		http.ServeFile(w, r, filepath.Join("public", "sample.txt"))
	})

	// ストリーミング（chunked。サーバ側から段階的に送る）
	mux.HandleFunc("/api/stream", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming unsupported", http.StatusInternalServerError)
			return
		}
		for i := 1; i <= 5; i++ {
			fmt.Fprintf(w, "chunk %d: %d\n", i, rand.Intn(1000))
			flusher.Flush()
			time.Sleep(600 * time.Millisecond)
		}
	})

	// アップロード（multipart/form-data）
	mux.HandleFunc("/api/upload", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		defer file.Close()
		// ここでは内容は捨てる（学習用）。サイズだけ数える
		n, _ := io.Copy(io.Discard, file)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"name": header.Filename,
			"size": n,
			"type": header.Header.Get("Content-Type"),
		})
	})

	// CORS（別オリジンから試す場合のための簡易実装）
	mux.HandleFunc("/api/cors", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "content-type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"ok": "cors"})
	})

	addr := ":18888"
	log.Printf("fetch api demo => http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
