package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // デモ用: 本番はオリジン制限を
	},
}

func replyFor(input string) string {
	// 前後の空白を除去して、全角スペースも削除
	trimmed := strings.TrimSpace(strings.ReplaceAll(input, "　", ""))

	switch trimmed {
	case "114":
		return "サーバちゃん「514"
	}

	// ちょっとだけ遊び: 「?」で終わるなら反射的に「いい質問！」で返す
	if strings.HasSuffix(trimmed, "?") || strings.HasSuffix(trimmed, "？") {
		return "いい質問！"
	}

	// デフォルトはエコー
	return fmt.Sprintf("サーバーちゃん 「%s", input)
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()

	log.Println("✅ クライアント接続")

	for {
		mt, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("read error:", err)
			break
		}
		in := string(msg)
		log.Printf("📩 受信: %q", in)

		out := replyFor(in)

		if err := conn.WriteMessage(mt, []byte(out)); err != nil {
			log.Println("write error:", err)
			break
		}
	}
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", wsHandler)
	mux.Handle("/", http.FileServer(http.Dir("./public")))

	log.Println("WebSocket demo => http://localhost:18888")
	http.ListenAndServe(":18888", mux)
}
