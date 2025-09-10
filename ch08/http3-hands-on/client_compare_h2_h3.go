// client_compare_h2_h3.go
package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/quic-go/quic-go/http3"
)

func run(label string, client *http.Client, base string) {
	start := time.Now()
	urls := []string{
		base + "/slow?d=500",
		base + "/slow?d=500",
		base + "/slow?d=500",
	}

	done := make(chan struct{}, len(urls))
	for _, u := range urls {
		go func(u string) {
			resp, err := client.Get(u)
			if err != nil {
				fmt.Println(label, "ERR:", err)
				done <- struct{}{}
				return
			}
			_, _ = io.ReadAll(resp.Body) // 読み切ることが大事
			resp.Body.Close()
			done <- struct{}{}
		}(u)
	}

	for i := 0; i < len(urls); i++ {
		<-done
	}
	fmt.Printf("[%s] total: %v\n", label, time.Since(start))
}

func main() {
	// --- HTTP/2 クライアント（自己署名証明書を許可） ---
	trH2 := &http.Transport{
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
		MaxConnsPerHost:     1, // 単一コネクションに縛る
		MaxIdleConnsPerHost: 1,
		ForceAttemptHTTP2:   true, // TLSなら自動でh2になる想定
	}
	clientH2 := &http.Client{Transport: trH2}

	// --- HTTP/3 クライアント ---
	h3t := &http3.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	defer h3t.Close()
	clientH3 := &http.Client{Transport: h3t}

	run("HTTP/2", clientH2, "https://localhost:8442")
	run("HTTP/3", clientH3, "https://localhost:8443")
}
