// server_h2_tls.go
package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

func slowH2(w http.ResponseWriter, r *http.Request) {
	dms := r.URL.Query().Get("d")
	ms, _ := strconv.Atoi(dms)
	if ms <= 0 {
		ms = 500
	}
	time.Sleep(time.Duration(ms) * time.Millisecond)
	fmt.Fprintf(w, "h2 slow done: %dms\n", ms)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/slow", slowH2)

	addr := ":8442"
	log.Println("HTTP/2 TLS server on https://localhost" + addr)

	// Go の TLS サーバは ALPN で自動的に HTTP/2 を有効化（証明書は自己署名でOK）
	if err := http.ListenAndServeTLS(addr, "server.crt", "server.key", mux); err != nil {
		log.Fatal(err)
	}
}
