世界の鍋圧が登場します

# HTTP/1.1 Upgrade（独自プロトコル）デモ README

> サーバーは `/upgrade` への **HTTP/1.1 Upgrade** を受けて 101 を返し、以降は **行単位テキストの独自プロトコル** に切り替えます。クライアントは Upgrade を要求し、101 以降はソケットをそのまま読み続けます。

---

## 1. リポジトリ構成（例）

```
03_protocol_upgrade/
├─ client_protocol_upgrade
│ 
├─ server_protocol_upgrade
│ 
└─ README.md
```

> ディレクトリ名は自由ですが、**server** と **client** を分けると試しやすいです。

---

## 2. 必要環境

* Go 1.20 以上（推奨: 1.22+）
* macOS / Linux / WSL2（Windows ネイティブでも可）

確認:

```bash
$ go version
```

---

## 3. まずは `go mod init`（一度だけ）

プロジェクト直下でモジュールを初期化します。

```bash
cd upgrade-demo
go mod init example.com/upgrade-demo
```

> 既に `go.mod` があるならこの手順は不要です。

---

## 4. サーバーの起動

サーバーコード（`server/main.go`）をそのまま使用します。ポートは `:18888` 固定。

```bash
# ターミナルA（サーバー用）
cd upgrade-demo/server
go run .
# もしくは: go run main.go
```

起動ログ例:

```
2025/09/23 21:40:12 listening on :18888
```

> **終了** は `Ctrl + C`。

---

## 5. クライアントの起動

別のターミナルでクライアントを実行します。

```bash
# ターミナルB（クライアント用）
cd upgrade-demo/client
go run .
# もしくは: go run main.go
```

出力例（先頭数行）:

```
2025/09/23 21:41:00 Status: 101 Switching Protocols
2025/09/23 21:41:00 Headers: map[Connection:[Upgrade] Upgrade:[MyProtocol]]
<- いち
<- にー
<- さん
<- よん
<- ごー
...
```

サーバー側ログ例:

```
2025/09/23 21:41:00 Upgrade to MyProtocol requested
2025/09/23 21:41:00 -> 1: いち
2025/09/23 21:41:00 -> 2: にー
2025/09/23 21:41:00 -> 3: さんぁああああああああああ
...
```

---

## 6. うまく動かない時のチェック

1. **ポート競合**: すでに :18888 を使っていませんか？

   * 対処: サーバーの `addr := ":18888"` を別ポート（例 `":19999"`）に変更して再起動。

2. **接続先ミス**: クライアントの接続先は `localhost:18888` になっていますか？

3. **ファイアウォール/企業VPN**: ローカルループバックに干渉するツールがないか確認。

4. **改行**: クライアントは `\n` 区切りで読みます。サーバー側が必ず `\n` を送っているか（`fmt.Fprintf(writer, "%s\n", out)`) を確認。

5. **Go のバージョン**: 古い Go だと `net/http` の挙動差で問題が出ることがあります。できれば 1.22+。

---

## 7. 仕組みの超ざっくり解説

### サーバー側

* `Connection: Upgrade` と `Upgrade: MyProtocol` を **要求された** リクエストだけ受け付けます。
* `http.Hijacker` で **HTTP レイヤを乗り捨て（Hijack）** し、素の TCP コネクションを取得。
* `101 Switching Protocols` を返したら **HTTP 終了**、以後は **任意のバイト列** を流せます（ここでは 200ms ごとに日本語の行を1..40まで）

### クライアント側

* 生のソケットに `GET /upgrade` を **自前で書き込み**、`Connection: Upgrade` / `Upgrade: MyProtocol` をセット。
* `http.ReadResponse` で **101 を確認**。それ以降は `bufio.Reader` で **行単位で受信** して表示。

> つまり、**ハンドシェイクだけHTTPを使い、以降は生ソケットで喋る** パターンの最小実装です。

---

## 8. cURL 等での確認メモ

cURL で 101 までは見られます（以降の独自プロトコルは人力観察向きではありません）。

```bash
curl -v \
  -H "Connection: Upgrade" \
  -H "Upgrade: MyProtocol" \
  http://localhost:18888/upgrade
```

> 101 が返れば成功。以降のストリームは cURL ではきれいに読めません。

---

## 9. バイナリ化して実行（任意）

```bash
# サーバー
cd upgrade-demo/server
go build -o upgrade-server
./upgrade-server

# クライアント
cd upgrade-demo/client
go build -o upgrade-client
./upgrade-client
```

---

## 10. よくあるQ\&A

**Q. サーバー・クライアントを同一プロセスで動かせる？**

* 実験なら可能ですが、学習用途では**別プロセス**にした方が挙動が分かりやすいです。

**Q. 101 の後も HTTP ヘッダは要る？**

* 要りません。**以降は HTTP ではない**ため、純粋にあなたの独自プロトコルの世界です。

**Q. keep-alive は？**

* Upgrade 以降は **アプリ側の責務**。ここではサーバーが 40 行流して終了。

---

## 11. 改造ポイント

* 送信上限 `1..40` を増やす/無限にする
* 送信間隔 `200ms` を調整
* 読みのルール（`readingJP` / `ahoTail`）を差し替える
* バイナリプロトコルに変えてみる（`\n` ではなく固定長ヘッダ+ペイロード等）


---


chatGPTにREADMEを書くようにお願いしたらすごいなこりゃ・・・