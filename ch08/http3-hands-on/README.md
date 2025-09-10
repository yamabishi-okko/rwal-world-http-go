OSI参照モデル ハンズオン 実験D
HTTP/2 と HTTP/3 の違いを体感する

この実験では、HTTP/2（TCPベース）と HTTP/3（QUIC = UDPベース）をGoで比較します。
両方に同じエンドポイント /slow?d=500 を用意し、同時3リクエストを投げて合計時間を比べます。

背景
HTTP/2
1本のTCP接続で多重化できる
ただしパケットロスがあると全体が足止め（TCPの仕組みによるHOL問題）

HTTP/3
TCPの代わりに QUIC (UDP) を使う
ストリームごとに独立しているので、1つのパケットロスが他に波及しない
モバイル環境やロスの多いネットワークで効果を発揮

## サーバ実装
HTTP/2（TLS）サーバ: server_h2_tls.go
mux := http.NewServeMux()
mux.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
    dms := r.URL.Query().Get("d")
    ms, _ := strconv.Atoi(dms)
    if ms <= 0 { ms = 500 }
    time.Sleep(time.Duration(ms) * time.Millisecond)
    fmt.Fprintf(w, "h2 slow done: %dms\n", ms)
})
http.ListenAndServeTLS(":8442", "server.crt", "server.key", mux)


HTTP/3（QUIC）サーバ: server_h3.go
mux := http.NewServeMux()
mux.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
    dms := r.URL.Query().Get("d")
    ms, _ := strconv.Atoi(dms)
    if ms <= 0 { ms = 500 }
    time.Sleep(time.Duration(ms) * time.Millisecond)
    fmt.Fprintf(w, "h3 slow done: %dms\n", ms)
})

cert, _ := tls.LoadX509KeyPair("server.crt", "server.key")
server := http3.Server{
    Addr:      ":8443",
    Handler:   mux,
    TLSConfig: &tls.Config{Certificates: []tls.Certificate{cert}},
}
server.ListenAndServe()



クライアント実装: client_compare_h2_h3.go
func run(label string, client *http.Client, base string) { ... }

trH2 := &http.Transport{
    TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
    MaxConnsPerHost:     1,
    MaxIdleConnsPerHost: 1,
    ForceAttemptHTTP2:   true,
}
clientH2 := &http.Client{Transport: trH2}

h3t := &http3.Transport{
    TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
}
defer h3t.Close()
clientH3 := &http.Client{Transport: h3t}

run("HTTP/2", clientH2, "https://localhost:8442")
run("HTTP/3", clientH3, "https://localhost:8443")


## 実行手順
1) 自己署名証明書を作成
openssl req -x509 -newkey rsa:2048 -nodes \
  -keyout server.key -out server.crt -days 365 \
  -subj "/CN=localhost" \
  -addext "subjectAltName=DNS:localhost,IP:127.0.0.1"

2) サーバを2つ起動
go run server_h2_tls.go
go run server_h3.go

3) クライアントで比較
go run client_compare_h2_h3.go


##　実行結果の例
[HTTP/2] total: 511ms
[HTTP/3] total: 504ms

ローカル環境では大きな差が出ないことも多い。
しかし、遅延やパケットロスを入れると HTTP/3 が安定して短時間で終わる傾向が見える。

## 学べること
HTTP/2 は TCP の上で動き、同時多重できるが「パケットロス時に全体が足止め」になる。
HTTP/3 は UDPベースの QUIC を使うことで、ストリームごとに独立しロスに強い。
実験では差が小さくても、ネットワークが悪いときに差が大きくなる。



