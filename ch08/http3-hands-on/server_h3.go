// server_h3.go
package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/quic-go/quic-go/http3"
)

func slowH3(w http.ResponseWriter, r *http.Request) {
	dms := r.URL.Query().Get("d")
	ms, _ := strconv.Atoi(dms)
	if ms <= 0 {
		ms = 500
	}
	time.Sleep(time.Duration(ms) * time.Millisecond)
	fmt.Fprintf(w, "h3 slow done: %dms\n", ms)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/slow", slowH3)

	// TLSの設定（自己署名証明書を使う）
	tlsConf, err := tls.LoadX509KeyPair("server.crt", "server.key")
	if err != nil {
		log.Fatal(err)
	}
	server := http3.Server{
		Addr:      ":8443",
		Handler:   mux,
		TLSConfig: &tls.Config{Certificates: []tls.Certificate{tlsConf}},
	}

	log.Println("HTTP/3 server on https://localhost:8443")
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}