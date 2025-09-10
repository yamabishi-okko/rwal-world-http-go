# 852 Server-Sent Events — サーバからの一方向ストリーム

このフォルダは **SSE (Server-Sent Events)** を最小構成で体験するためのデモです。  
ブラウザは接続を **張りっぱなし** にしておき、サーバはイベントを **流し続け** ます（1方向）。

> 目的：Fetch（都度リクエスト）との違い、HTTP 上での「継続的な配信」を体感する。

---
## 前提
- Go 1.20+（1.22 推奨）
- モダンブラウザ（Chromium / Firefox / Safari いずれも OK）
- 構成:
server.go # SSE エンドポイント /events と静的ファイル配信
public/
index.html # 接続ボタンと表示領域
main.js # EventSource を使って /events に接続
---

## 起動

```bash
cd ch08/85_js_comms/852_sse
go run .
# => SSE demo => http://localhost:18888

ブラウザで http://localhost:18888 を開き、ボタンを押すと
1秒おきに「メッセージ 1〜5」 が流れてきます（このデモは5件で終了）。

##仕組み（吹き出し図）
💻 ブラウザ (index.html + main.js)
「EventSource('/events') でつなぎっぱなしにするよ！」
         ├── 接続を張る（HTTP/1.1 keep-alive / HTTP/2 ストリーム）
         ▼
🖥 サーバ (server.go)
「OK！ 'text/event-stream' で1秒ごとに data を流すね〜」
data: メッセージ 1
data: メッセージ 2
…（空行で区切るのが SSE のルール）
         ▲
         └── ブラウザ側は onmessage で逐次受け取って表示
レスポンスヘッダは必ず Content-Type: text/event-stream。

イベントは data: ... 行 を書いて 空行 で区切る（これが1イベント）。


## 画面での見方
ボタンを押す → 下部の <pre> に
🌙🐚ソウセージ🪼☁️ 1
…
🌙🐚ソウセージ🪼☁️ 5
エラー: [object Event]
のように追記されます。
※ 最後の「エラー」は 接続が閉じられた通知（このデモは5件送って閉じる実装）。
DevTools → Network → /events を選ぶ
Headers: Content-Type: text/event-stream, Cache-Control: no-cache
Timing/Waterfall: 転送が継続している（完了にならない）



##イベントのフォーマット（拡張）
SSE の行は複数種類があります。よく使うのは次の4つ：
id: 11
event: ping
data: {"time":"2025-09-09T00:00:00Z"}
retry: 3000

<空行で1イベント終わり>

data: … 本体（複数行 OK。複数行の場合は連結される）
event: … 名前付きイベント。JS 側は evtSource.addEventListener('ping', ...)
id: … 受信側が最後の ID を覚えて、再接続時に Last-Event-ID ヘッダで再開要求
retry: … 自動再接続までのミリ秒（ブラウザが尊重）
このデモは最小化のため data: のみを送信しています。発展として event:/id: を追加してみてください。


##コードの読み方（最小ポイント）
server.go（要点）
/events のハンドラでヘッダを設定：
Content-Type: text/event-stream
Cache-Control: no-cache
Connection: keep-alive
w.(http.Flusher) を使って Flush（チャンク化して逐次送信）
サンプルは 5回ループして送信／最後に接続が閉じる
public/main.js（要点）
new EventSource('/events') で接続
onmessage = (e) => { ... } で受け取り
onerror = () => { ...; close(); } で終了を検知（このデモはサーバ側が閉じるのでここに入る）


##Fetch / WebSocket との違い
|   　 　　| Fetch          | **SSE**                            | WebSocket               |
| ----　　 | -------------　| ---------------------------------- | ----------------------  |
| 通信形態　| リクエスト都度   | **サーバ→クライアント一方向**          | 双方向（フルデュプレックス） |
| 接続     | 毎回張る 　    　| **張りっぱなし**（HTTP 上のストリーム） | Upgrade して別プロトコル   |
| 用途     | API 取得・送信   | 通知/進捗/ログ/一方向の定期配信        | チャット/ゲーム/双方向同期   |
| 対応     | どのブラウザもOK | 主要ブラウザ対応                      | 主要ブラウザ対応           |


「一方向で十分・実装を軽くしたい」なら SSE はとても良い選択です。
よくあるつまずき
プロキシ/リバースプロキシでバッファされてしまう
Nginx なら proxy_buffering off; が必要な場合があります。
まずは 直アクセス（localhost → アプリ直）で挙動確認を。
CORS でブロック
別オリジンのページから接続する場合は、サーバで Access-Control-Allow-Origin を付与。
接続がすぐ切れる
コンテナ/ロードバランサのアイドルタイムアウトに注意。
サーバ側で定期的にコメント行（: keepalive\n\n）を送ると維持しやすい。


用語メモ
SSE / EventSource：HTTP 上の一方向ストリーミング。MIME は text/event-stream。
Flusher：Go の http.Flusher。レスポンスを即座にソケットへ流す。
id / event / data / retry：SSE の行フォーマット。空行で1イベント終端。
Keep-Alive：接続を維持して同一レスポンスを流し続ける。



✅ SSE のメリット
HTTP だけで動く（実装がシンプル）
WebSocketみたいに「Upgrade」や独自プロトコルを扱う必要がない。
普通の HTTP サーバ (text/event-stream) を書ければすぐ使える。
ブラウザ標準の API がある
EventSource を new するだけで、再接続やエラー処理まで含めて動く。
特別なライブラリを追加しなくても、モダンブラウザならほぼ対応済み。
自動再接続が標準装備
接続が落ちてもブラウザが勝手に再接続してくれる。
id: と Last-Event-ID を使えば「切れたところから再開」もできる。
一方向通信に最適化
サーバから「プッシュ型通知」を送りたい場合にちょうど良い。
例: チャットの新着通知、株価の更新、ログ/メトリクス配信など。
HTTP/2 / HTTP/3 と相性が良い
1本のコネクションで複数ストリームを持てるため、大量のSSE接続も効率的にさばける。
WebSocket だとコネクションごとに管理が必要だが、SSEはHTTPの枠内で完結。

❌ 制約（デメリット）
クライアント → サーバ には送れない（サーバ → クライアント専用）。
バイナリ非対応（テキストのみ。必要ならBase64で送る）。
プロキシやロードバランサにバッファされやすい（設定次第では一括配送になってしまう）。

🌙まとめ
SSE の最大のメリットは、
👉 「サーバから一方向に流しっぱなしで送りたいとき、WebSocketより圧倒的にお手軽」という点！
通知・ログ・ストリーム表示ならまず SSE が候補、
双方向チャットやゲームのようなケースなら WebSocket へ、
という住み分けになるよ。
