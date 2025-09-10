#  トランスポート層

README.md（実験B: TCPソケットでHTTPSアクセス）
実験B: トランスポート層＋アプリケーション層を手書きで体験
HTTPSサイトに TCP + TLS でアクセス

この実験では、GoでTLS付きTCPソケットを直接操作し、HTTPSサイトにアクセスします。
HTTPライブラリは使わず、自分でリクエスト行を手書きして送信します。

コード解説（tcp_manual_https.go）
package main

## TLSで接続する
    conn, err := tls.Dial("tcp", "www.poplar.co.jp:443", nil)
    if err != nil {
        panic(err)
    }
    defer conn.Close()
tls.Dial … TCP + TLSでサーバー（ポプラ社の443番ポート）に接続する。
err … 接続に失敗したらここにエラーが入る。
defer conn.Close() … 最後に必ず接続を閉じる。

## HTTPリクエストを自分で書く
    req := "" +
        "GET /zorori/ HTTP/1.1\r\n" +
        "Host: www.poplar.co.jp\r\n" +
        "User-Agent: go-osi-hands-on/1.0\r\n" +
        "Accept: */*\r\n" +
        "Connection: close\r\n" +
        "\r\n"
GET /zorori/ HTTP/1.1 … リクエスト行。/zorori/ ページを取る命令。
Host: www.poplar.co.jp … アクセス先のサーバー名。HTTP/1.1では必須。
User-Agent: ... … クライアント情報。今回は自作の文字列。
Accept: */* … どんなコンテンツでも受け取る。
Connection: close … レスポンスが終わったら接続を切る。
\r\n … HTTPヘッダの改行。最後は空行で区切る必要がある。

## リクエスト送信
    if _, err := fmt.Fprint(conn, req); err != nil {
        panic(err)
    }
fmt.Fprint(conn, req) … ソケットにリクエスト文字列を書き込む。
実際には暗号化されてTLSで送られる。

## レスポンスを受け取る
    if _, err := io.Copy(os.Stdout, conn); err != nil && !errors.Is(err, io.EOF) {
        panic(err)
    }
}
io.Copy(os.Stdout, conn) … サーバーの返答（ヘッダ＋本文）を全部ターミナルに出力する。
io.EOF は「もうデータがありません」という終了合図なのでエラーではない。


## 実行方法
cd osi-hands-on
go run tcp_manual_https.go


## 結果の例
HTTP/1.1 200 OK
Date: Thu, 12 Sep 2025 04:30:00 GMT
Content-Type: text/html; charset=UTF-8
Content-Length: ...
Connection: close
...
<!DOCTYPE html>
<html lang="ja">
<head>...</head>
<body>
  ...（ゾロリ公式ページのHTML）...
</body>
</html>


## 学べること
HTTPSの正体は「TCP + TLS + HTTP」。
net/http を使わなくても、ソケットにHTTPリクエストを文字列で書けば同じように動く。
リクエストの基本構造（リクエスト行 → ヘッダ群 → 空行）と、レスポンスの構造（ステータス行 → ヘッダ群 → 空行 → 本文）が体感できる。