// tcp_manual_https.go
package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"os"
)

func main() {
	// 1) トランスポート層: TLS付きTCP接続を開く (443番ポート)
	conn, err := tls.Dial("tcp", "www.poplar.co.jp:443", nil)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// 2) アプリ層: HTTPリクエスト行＋ヘッダを書き込む
	req := "" +
		"GET /zorori/ HTTP/1.1\r\n" +
		"Host: www.poplar.co.jp\r\n" +
		"User-Agent: go-osi-hands-on/1.0\r\n" +
		"Accept: */*\r\n" +
		"Connection: close\r\n" +
		"\r\n"

	if _, err := fmt.Fprint(conn, req); err != nil {
		panic(err)
	}

	// 3) サーバの返答をそのまま標準出力に流す
	if _, err := io.Copy(os.Stdout, conn); err != nil && !errors.Is(err, io.EOF) {
		panic(err)
	}
}
