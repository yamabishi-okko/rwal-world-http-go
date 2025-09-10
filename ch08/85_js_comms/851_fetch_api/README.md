# 851 Fetch API — ブラウザからの基本HTTP通信をぜんぶ試す

このフォルダは、**Fetch API を使ってブラウザから HTTP の典型パターンを実験する最小セット**です。  
8.5.1（Fetch API）の内容に沿って、以下をひとつのサーバで確認できます。

- JSON の取得（GET）
- サーバからの **ストリーミング送信**（chunked transfer）
- ブラウザ強制ダウンロード（`Content-Disposition: attachment`）
- ファイルアップロード（`multipart/form-data`）
- CORS の最小挙動

> 目的：HTTP/2/HTTP/3 へ進む前に、**HTTP/1.1 でも体験できる通信の型**を手で動かして感覚を掴む。

---

## 前提

- Go 1.20+（1.22 推奨）
- モダンブラウザ（Chromium 系 / Firefox いずれか）
- このフォルダにあるファイル：
- server.go # Go製の超ミニサーバ
public/
index.html # 操作用UI（ボタンいくつか）
main.js # Fetch の呼び出しコード
sample.txt # ダウンロード用のテキスト


---

## 起動

###  このフォルダから起動（推奨）
```bash
cd ch08/85_js_comms/851_fetch_api
go run .



使い方（UI で試す）

画面に並ぶボタンを押すと、それぞれのエンドポイントを叩きます。結果は画面下部の <pre> に表示されます。

GET /api/hello

JSON を返すだけの最小 API。

DevTools → Network → Response/Headers で
Content-Type: application/json; charset=utf-8 を確認。

GET /api/stream（ストリーミング）

サーバが 0.6 秒おきに chunk 1..5 を送信。

ReadableStream 経由で段階的に描画されます。

Network → Timing で転送が分割されている様子が見えます。

企業プロキシ経由だと一括で届く場合あり（後述の「注意」参照）。

Download（強制ダウンロード）

Content-Disposition: attachment; filename="sample.txt" により保存ダイアログへ。

Network → Response Headers でヘッダを確認。

Upload（multipart/form-data）

ファイルを選び Upload。サーバは FormFile("file") で受け取り、
name/size/type を JSON で返します。

Network → Request Payload を開くと multipart/form-data; boundary=... が見えます。

CORS の最小確認（/api/cors）

サーバは Access-Control-Allow-Origin: * を返す簡易実装。

別オリジンのページからも呼べます（下の「発展」参照）。




# JSON
curl -i http://localhost:18888/api/hello

# ストリーミング（-N: バッファせず逐次表示）
curl -N http://localhost:18888/api/stream

# ダウンロード（-O で保存, -J でサーバ指定名を採用 → sample.txt）
curl -OJ http://localhost:18888/api/download

# アップロード（任意のファイルでOK）
curl -i -F "file=@README.md" http://localhost:18888/api/upload

# CORS（Origin を明示して叩く）
curl -i -H "Origin: http://example.com" http://localhost:18888/api/cors




各エンドポイントの仕様（要点だけ）
| Path            | Method      | 返すもの / 受け取るもの                            　　 | 重点ポイント                                  
| --------------- | ----------- | --------------------------------------------------- | ------------------------------------------------------  |
| `/api/hello`    | GET         | `{"message": "...","time":"RFC3339"}`               | 最小の JSON API                                          |
| `/api/stream`   | GET         | `text/plain` を **5 チャンク**に分けて送信     　       | `http.Flusher` による chunked 送出 / ReadableStream 受信 　|
| `/api/download` | GET         | `sample.txt` を attachment として返す                 | `Content-Disposition`                      　    　　　　 |
| `/api/upload`   | POST        | `multipart/form-data`（フォーム名 `file`）→ JSON 返却  | サーバ側の `FormFile` で受理    　                     　   |
| `/api/cors`     | GET/OPTIONS | 200（`Access-Control-Allow-Origin: *`）         　   | 簡易 CORS                             　　　              |


このサーバは HTTP/1.1 で動作します。HTTP/2/3 では「フレーム化」など下位表現は変わりますが、アプリ層の見かけは同じです（ストリーミングや優先度の体感が次章で効いてきます）。


観察ポイント（DevTools で見るべき所）

Headers

download の Content-Disposition

upload の Content-Type: multipart/form-data; boundary=...

cors の Access-Control-Allow-Origin: *

Timing

stream が 少しずつ届く（Waterfall に段差）

Preview / Response

stream はプレビューが徐々に伸びる

Request

upload の各パート（Content-Disposition: form-data; name="file"; filename="..."）

学べること（このフォルダの存在意義）

Fetch API で HTTP の基本（GET/POST/ヘッダ） を把握

chunked ストリーミングを体感（→ HTTP/2/3 の「並行・フレーム」の理解に種を蒔く）

ブラウザの挙動を制御するヘッダ（例: Content-Disposition）を把握

multipart/form-data の実体を目で見る（JSON 以外のボディ形式の理解）

CORS の最小形（許可ヘッダがないとブラウザがブロックする理由）


## このフォルダの構成とファイル同士のつながり

このデモは **Go サーバ (`server.go`)** と **ブラウザで開くフロント (public/ 内)** の 2 部構成です。

### server.go
- Go 言語で書かれた簡易 HTTP サーバ。
- 役割は次の 2 つ：
  1. `/` にアクセスしたら **public/index.html** を返す  
  2. `/api/...` にアクセスしたら API 用の処理を返す（JSON やストリームなど）

つまり **「ページを見せる」係** と **「ボタンを押した時に応答する」係** を一つにまとめたものです。

### public/index.html
- ブラウザで最初に開くページ。
- 中には「GET /api/hello」「GET /api/stream」などのボタンが置いてある。
- ボタンを押すと **main.js の関数** が呼ばれる。

### public/main.js
- JavaScript で書かれたクライアント側の処理。
- `fetch()` を使ってサーバの `/api/...` にリクエストを送る。
- サーバから返ってきたレスポンスを受け取り、画面に表示する。
- 例：
  - `fetch('/api/hello')` → server.go の `/api/hello` が動いて JSON を返す
  - `fetch('/api/stream')` → server.go がチャンク送信し、それを `ReadableStream` で表示

### public/sample.txt
- ダウンロードのデモに使うサンプルファイル。
- サーバの `/api/download` を叩くと、`Content-Disposition: attachment` ヘッダ付きでこのファイルが送られてくる。
- その結果、ブラウザは自動的に「保存ダイアログ」を開く。

---

## 処理の流れ（例: 「GET /api/hello」ボタンを押したとき）

1. ユーザがブラウザでボタンをクリック
2. `index.html` に書いてあるイベントが `main.js` を呼ぶ
3. `main.js` 内で `fetch('/api/hello')` が実行される
4. リクエストが `server.go` に届く
5. `server.go` が `{"message":"hello","time":"..."}` を JSON で返す
6. `main.js` が結果を受け取り、画面の `<pre>` に表示

---

このように、
- **server.go** … サーバ側の処理（API + 静的ファイル配信）  
- **index.html / main.js** … フロント側の処理（UI + fetchでAPI呼び出し）  
- **sample.txt** … デモ用の静的ファイル  

が役割分担してつながっています。

## ファイル同士のつながり（図解）

ブラウザでボタンを押してから画面に結果が出るまでの流れを図にしました。

💻 ブラウザ (index.html + main.js)
「ボタン押されたよ！ fetch('/api/hello') で呼ぶね！」
│
▼
｜リクエスト送信→
│
▼
🖥 サーバ (server.go)
「OK！ /api/hello に来たね。JSON を返すよ〜」
{
"message": "hello from fetch api",
"time": "2025-09-08T23:59:59Z"
}
│
▼
←レスポンス返却｜
│
▼
💻 ブラウザ
「JSON 受け取った！ 画面に表示するね ✨」
│
▼
📂 sample.txt
「私はただのファイルだけど、
server.go に呼ばれたら
ダウンロード用に届けられるよ！」


---
### ポイント
- **ブラウザ側** → index.html のボタン / main.js の fetch がセリフ担当  
- **サーバ側** → server.go が返事役  
- **sample.txt** → 「呼ばれたら渡されるファイル」というキャラで登場  
---

