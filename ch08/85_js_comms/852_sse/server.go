package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	mux := http.NewServeMux()

	// SSEエンドポイント
	mux.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		// ヘッダをSSE用にセット
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		// Flusherで逐次送信を有効化
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
			return
		}

		// 5回だけメッセージ送信
		for i := 1; i <= 5; i++ {
			fmt.Fprintf(w, "data:🌙🐚ソウセージ🪼☁️ %d\n\n", i)
			flusher.Flush()
			time.Sleep(1 * time.Second)
		}
	})

	// 静的ファイル (index.html, main.js)
	mux.Handle("/", http.FileServer(http.Dir("./public")))

	log.Println("未納金に関する詳細をすぐクリックして確認 => http://localhost:18888")
	http.ListenAndServe(":18888", mux)
}
