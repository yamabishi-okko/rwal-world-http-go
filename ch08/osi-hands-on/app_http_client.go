// app_http_client.go
package main

import (
	"fmt"
	"io"
	"net/http"
)

func main() {
	// アプリ層：HTTPのGET（下のTCPはライブラリがよしなにやってくれる）
	resp, err := http.Get("https://www.poplar.co.jp/zorori/")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("Status:", resp.Status) // ← ここでアプリ層の「結果（ステータス）」を確認
	body, _ := io.ReadAll(resp.Body)
	fmt.Println(string(body))           // ← これが「コンテンツ（本文）」
}
