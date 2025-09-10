#　アプリケーション層

OSI参照モデル ハンズオン 実験A
アプリケーション層（HTTPクライアント）

この実験では、**OSI参照モデルの「アプリケーション層」**をGoで体験します。
Goの標準ライブラリ net/http を使って、ゾロリ公式 にアクセスし、HTTPレスポンスを受け取ります。

コード解説（app_http_client.go）


## package main
Goプログラムは必ず最初に package を書く。
main パッケージは「このファイルが実行可能プログラムですよ」という意味。
## import (
    "fmt"
    "io"
    "net/http"
)

import：必要なライブラリを読み込む。
fmt … 文字出力用（print）。
io … 入出力の読み書き。
net/http … HTTP通信のライブラリ。

## func main() {

    main関数。Goではここからプログラムが動き始める。
    resp, err := http.Get("http://example.com")
    http.Get … 指定URLにHTTP GETリクエストを送る。
    resp … サーバーから返ってきたレスポンスが入る。
    err … 通信に失敗したらここにエラーが入る。

##     if err != nil {
        panic(err)
    }

エラーがあれば panic で強制終了（今回は簡単のため）。

##     defer resp.Body.Close()
通信が終わったら必ず レスポンスの中身を閉じる。
defer は「この関数が終わるときに実行してね」という予約。

##      fmt.Println("Status:", resp.Status)
サーバーからの返答の「ステータスコード」を表示。
例：200 OK（成功）、404 Not Found（ページなし）など。

##     body, _ := io.ReadAll(resp.Body)
    fmt.Println(string(body))
resp.Body … サーバーから返ってきた本文（HTML）。
io.ReadAll … 本文を全部読み込む。
string(body) … バイトデータを文字列に変換して表示。
  

  # 実行方法
  cd osi-hands-on
go run app_http_client.go

# 学べること
OSI参照モデルのアプリケーション層（HTTP）を直接体験できる。z
HTTPは「URLを指定→サーバーにリクエスト→ステータス＋本文が返る」という流れ。
下の層（TCPやIP）は全部 net/http が隠してくれるので、アプリ層にだけ集中できる。

