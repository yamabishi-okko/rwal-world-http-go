HTTP/1.1 vs HTTP/2 を体感する（ゾロリ公式で比較）
目的
同じURLに同時3リクエストを投げ、合計時間の違いから
HTTP/1.1（直列的になりやすい） と HTTP/2（同時多重で並行） の体感差を理解する。
あなたの実測例
[HTTP/1.1] total: 1.343235833s

[HTTP/2] total: 216.842834ms

→ HTTP/2 のほうが圧倒的に速い（同時多重の効果）
注意：外部サイトに負荷をかけないため、同時3本×1回だけに留めます。

前提
Go 1.20+（目安）
モジュール初期化＆依存追加（最初の一回だけ）

cd ~/play/1.0x2.0              # client_compare_zorori.go のあるフォルダ
go mod init client-compare-zorori
go get golang.org/x/net/http2


## 実行方法
go run client_compare_zorori.go

## 仕組みの違い
HTTP/1.1（単一接続・直列になりやすい）
┌──────────── TCP接続(1本) ────────────┐
   [Req1]───待つ───[Res1] → [Req2]───待つ───[Res2] → [Req3]───待つ───[Res3]
└──────────────────────────────────────┘
合計 ≒ 処理時間の合計（ヘッド・オブ・ライン・ブロッキングが起きやすい）

HTTP/2（単一接続・同時多重）
┌──────────── TCP接続(1本) ────────────┐
   [Req1]  [Req2]  [Req3]   ← 同じ線路の中で“フレーム”を細切れに多重化
   [Res1]  [Res2]  [Res3]   ← 並行で返せる
└──────────────────────────────────────┘
合計 ≒ 最も遅い1本の時間（最大レイテンシ）


### HTTP/1.1
Keep-Alive で同じ接続を使い回せるが、同時多重は不可。
パイプラインという“先行送信”は仕様上あるが、実運用では問題が多く一般的に無効。
その結果、直列処理になりやすい → 合計時間が積み上がる。

### HTTP/2
バイナリフレーミングとストリームの多重化で、1本の接続上で同時に複数リクエストを流せる。
ヘッダは HPACK で圧縮。
優先度やサーバープッシュの仕組みもあるが、プッシュは現在ほとんど使われないことが多い。

補足：HTTP/2 は1本の TCP 上で多重化するため、TCPのパケット損失があると全体に影響し得ます（TCPレベルのHOL）。この弱点は HTTP/3（QUIC） で改善されます。


## コード解説（client_compare_zorori.go）
1) 依存ライブラリ
net/http … HTTPクライアントの標準ライブラリ
golang.org/x/net/http2 … HTTP/2 を有効化するために使用

2) 比較対象のURL
const target = "https://www.poplar.co.jp/zorori/"

ゾロリ公式の同一URLに 3 本並行アクセスします。

3) run関数：同時3本の合計時間を測る
直列ではなく goroutine で 3 本を同時に開始
io.ReadAll(resp.Body) で本文を消費（読み切らないと計測が不正確になる）
3本すべて終了したタイミングで elapsed を計測して表示
ポイント： 同時に開始しているのに、HTTP/1.1 クライアントでは直列的な挙動になり、合計時間が伸びやすい。
HTTP/2 クライアントでは合計が短縮され、最も遅い1本にほぼ一致する。

4) main関数：2種類のクライアントを作る
HTTP/1.1 クライアント
Transport を単一接続に縛る（純度の高い比較のため）
TLSNextProto を空マップにして、Go の自動 HTTP/2 アップグレードを無効化
trH1 := &http.Transport{
    MaxConnsPerHost:     1,
    MaxIdleConnsPerHost: 1,
    TLSNextProto:        map[string]func(string, *tls.Conn) http.RoundTripper{},
}
clientH1 := &http.Client{Transport: trH1}

HTTP/2 クライアント
同じく単一接続に縛る
http2.ConfigureTransport(trH2) で ALPN により h2 を有効化
trH2 := &http.Transport{
    MaxConnsPerHost:     1,
    MaxIdleConnsPerHost: 1,
}
_ = http2.ConfigureTransport(trH2)
clientH2 := &http.Client{Transport: trH2}
 

## 何が速くなるの？
1) 接続の使い回し
h1 でも Keep-Alive はあるが、1本の接続で同時に複数の応答は返せない。
h2 は 同一接続内で並行（多重化） できるので、急に強い。
2) ヘッダ圧縮（HPACK）
似たヘッダが繰り返し現れても、h2 は差分を効率よく送れる。
今回のように同一URLを連続で叩くケースでは効果が出やすい。
3) フレーミング
h2 はデータを細かいフレームに切って交互に流すので、1つの重いレスポンスが他をブロックしにくい。

### よくあるつまずき
1) h2になっていない
サーバ側が h2 非対応の経路だと h1 になる。http2.ConfigureTransport は「使えれば h2」という意味。
2) キャッシュや超高速回線で差が小さい
数回実行して平均を見ると傾向が見えやすい。
3) 会社やプロキシ環境
経路の制約で h2 が有効にならないことがある。

## まとめ
HTTP/1.1 は直列になりやすく、合計時間が積み上がる（HOL問題）。
HTTP/2 は1本の接続を同時多重でき、合計時間が短くなる。
Go では Transport の設定だけでこの差を再現・体感できる。
さらに先へ：HTTP/3（QUIC） は TCP の制約を超えて、パケット損失時の足止めを避けられる。