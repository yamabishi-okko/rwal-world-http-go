package main

import (
	"io"
	"log"
	"net/http"
)

func main() {
	// GET リクエスト送信
	resp, err := http.Get("http://localhost:18888")
	if err != nil {
		panic(err)
	}
	//終わったら閉じる
	// //ステータス表示　（文字列or数値）
	log.Println("Status:", resp.Status)
	log.Println("StatusCode:", resp.StatusCode)
	log.Println("Fields:", resp.Header)
	defer resp.Body.Close()

	// リクエストボディーに埋め込み
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
		
	///コンテンつ表示
	log.Println("Body:", string(body))



    // // ヘッダー一覧を表示
    // log.Println("Headers:", resp.Header)

    // // 特定のヘッダーを取り出す
    // log.Println("Content-Length:", resp.Header.Get("Content-Length"))
}