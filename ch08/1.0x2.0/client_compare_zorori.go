// client_compare_poplar.go
package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"time"

	"golang.org/x/net/http2"
)

const target = "https://www.poplar.co.jp/zorori/"

// run 同時に3本叩いて合計時間を計測
func run(label string, client *http.Client) {
	start := time.Now()
	n := 3

	done := make(chan struct{}, n)
	for i := 0; i < n; i++ {
		go func() {
			resp, err := client.Get(target)
			if err != nil {
				fmt.Println(label, "ERR:", err)
				done <- struct{}{}
				return
			}
			_, _ = io.ReadAll(resp.Body)
			resp.Body.Close()
			done <- struct{}{}
		}()
	}

	for i := 0; i < n; i++ {
		<-done
	}
	elapsed := time.Since(start)
	fmt.Printf("[%s] total: %v\n", label, elapsed)
}

func main() {
	// ★ HTTP/1.1 クライアント（単一コネクションに縛る / HTTP2を明示的に無効化）
	trH1 := &http.Transport{
		// 1コネクションに制限：並列要求を直列化させる
		MaxConnsPerHost:     1,
		MaxIdleConnsPerHost: 1,
		// GoはTLSだと自動でHTTP/2に上げようとするので、ここで無効化
		TLSNextProto: map[string]func(string, *tls.Conn) http.RoundTripper{},
		// （ダイヤラはデフォルトでOK）
	}
	clientH1 := &http.Client{Transport: trH1}

	// ★ HTTP/2 クライアント（h2を有効化）
	trH2 := &http.Transport{
		// 1コネクションに制限（h2はこの1本を多重化で使いまわす）
		MaxConnsPerHost:     1,
		MaxIdleConnsPerHost: 1,
	}
	// HTTP/2を有効化（ALPNでh2が使えればh2に）
	_ = http2.ConfigureTransport(trH2)
	clientH2 := &http.Client{Transport: trH2}

	// 実行（順序はどちらでもOK）
	run("HTTP/1.1", clientH1)
	run("HTTP/2", clientH2)
}
