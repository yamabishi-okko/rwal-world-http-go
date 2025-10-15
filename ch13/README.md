**「HTTPって結局なにしてるの？」を手で動かして感じるための超ミニ実験です。<br>
むずかしい言葉は最小限。まずは動かす → 触る → 見るの順でOK。🌙**

これで何ができる？（超ざっくり）<br>
・ボタンを押すとサーバからJSONやテキストが返ってくる<br>
・Cookieを使って「同じ人だよね？」を見分ける<br>
・URLの「?q=go」みたいなクエリや「/user/123」みたいなパスの値を受け取る<br>
・ぜんぶ Goの標準ライブラリだけ で実装<br>

### 起動手順
フォルダに移動<br>
```cd rwal-world-http-go/ch13```

初回だけ（まだなら）モジュール初期化<br>
```go mod init example.com/ch13```

起動<br>
```go run .```

### 触り方（まずはブラウザ） 
### 開く → http://localhost:8080/static/index.html 
　画面のボタンが fetch('/xxx') を呼ぶ<br>
　ブラウザが HTTPリクエスト を作って OS に投げる → TCPで localhost:8080 に到達<br>
　Go の net/http が受け取り、ServeMux（ディスパッチャ）がURLでハンドラに振り分け<br>
　ハンドラが処理して、ステータス・ヘッダ・ボディ を組み立て → HTTPレスポンス を返す<br>
　ブラウザがレスポンスを受け取り、fetch() の Promise が解決 → JSでJSONにし、preへ表示<br>

```
<button id="btn-time">/json をFetch</button>
<pre id="out"></pre>
<script>
  const out = document.getElementById('out');
  async function go(path) {
    const r = await fetch(path);                 // ← HTTPリクエストを送信
    const maybeJson = await r.json().catch(()=>({text:'not json'}));
    out.textContent = r.status + " " + r.statusText + "\n" +
                      JSON.stringify(maybeJson, null, 2);  // ← 画面に描画
  }
  document.getElementById('btn-time').onclick  = () => go('/json');
</script>
```
 
### GET /json … 現在時刻つきのJSONが表示される 
 原理：<br>
  レスポンスヘッダに Content-Type: application/json を付ける<br>
  json.Encoder が Goの値（構造体/マップ）→ JSON文字列へ変換し、ボディに直書き<br>
  ステータスはデフォルト 200 OK（変えたいなら w.WriteHeader(201) などを先に呼ぶ）<br>
 ブラウザ側：<br>
  fetch の Response を r.json() で逆変換してJSオブジェクトへ<br>
  それを JSON.stringify して pre>に表示<br>
     ```
    func handleJSON(w http.ResponseWriter, r *http.Request) {
        type resp struct{ Message, Time string }
        writeJSON(w, resp{Message: "hi", Time: time.Now().UTC().Format(time.RFC3339)})
    }

    func writeJSON(w http.ResponseWriter, v any) {
        w.Header().Set("Content-Type", "application/json; charset=utf-8") // ← ヘッダ
        enc := json.NewEncoder(w)                                         // ← エンコーダ
        enc.SetIndent("", "  ")
        _ = enc.Encode(v)                                                 // ← ボディにJSONを書き込む
    }
     ```

### GET /set-cookie … サーバがあなた用のIDを発行（Cookieに保存） 
 原理：<br>
  サーバが新規セッションIDを作って、サーバ側メモリMapにひも付いたデータを保存<br>
  レスポンスで Set-Cookie: sid=... を送る（これがブラウザに保存される）<br>
 ブラウザ側：<br>
  同一オリジンなら、次のリクエストから自動で Cookie: sid=... を付けて送る（fetchの既定動作）<br>
    ```
    func handleSetCookie(w http.ResponseWriter, r *http.Request) {
        sid := newSID()
        sessionsMu.Lock()
        sessions[sid] = &sessionData{Counter: 1, Created: time.Now()}
        sessionsMu.Unlock()

        http.SetCookie(w, &http.Cookie{
            Name: "sid", Value: sid, HttpOnly: true, SameSite: http.SameSiteLaxMode, Path: "/",
        })                                           // ← ★ レスポンスヘッダ Set-Cookie を出す
        writeJSON(w, map[string]any{"session": sessions[sid]})
    }
    ```

### GET /whoami … さっきのIDで「同じ人だね」と認識、カウンタが増える 
 原理：<br>
  ブラウザが勝手に付けてきた Cookie: sid=... を受け取り、サーバ側Mapのキーとして照合<br>
  見つかれば同一人物と判定し、カウンタを+1 → JSONで返す<br>
  見つからなければ 401 Unauthorized を返す（状態がないので“誰かわからない”）<br>
    ```
    func handleWhoAmI(w http.ResponseWriter, r *http.Request) {
        c, err := r.Cookie("sid")                 // ← ★ リクエストヘッダの Cookie から取り出す
        if err != nil { http.Error(w,"no session", http.StatusUnauthorized); return }
        sessionsMu.Lock(); defer sessionsMu.Unlock()
        s, ok := sessions[c.Value]
        if !ok { http.Error(w,"no session", http.StatusUnauthorized); return }
        s.Counter++                               // ← サーバ側の状態を更新
        writeJSON(w, map[string]any{"session": s})
    }
    ```

### GET /search?q=go … q に入れた文字をサーバが受け取って返す 
 原理：<br>
  r.URL.Query() が map[string][]string（クエリの辞書）を返す<br>
  それをそのままJSONでエコー<br>
    ```
    func handleSearch(w http.ResponseWriter, r *http.Request) {
    q := r.URL.Query().Get("q")   // ← URLの ?q= の値を取り出す
    writeJSON(w, map[string]string{"q": q})
    }
    ```

### GET /user/123 … URLの 123 をサーバが読み取って返す 
 原理：<br>
  最小実装なので自分で文字列を切る（本格ルータならパラメータ抽出をやってくれる）<br>
  バリデーションOKならJSON返却、NGなら 400 Bad Request<br>
    ```
    func handleUser(w http.ResponseWriter, r *http.Request) {
    id := strings.TrimPrefix(r.URL.Path, "/user/") // ← パスから手動で切り出し
    if _, err := strconv.Atoi(id); err != nil {
        http.Error(w, "id must be number", http.StatusBadRequest)
        return
    }
    writeJSON(w, map[string]string{"id": id})
    }
    ```

見どころ：ブラウザの Networkタブ を開いて、 「ステータス（200や401）」「ヘッダ（Set-Cookie / Cookie）」を覗いてみるだけで理解が進むよ。<br>


6. ひとつだけ“端から端まで”追ってみる（/json）
クリック → JS が fetch('/json')
ブラウザが GET /json を localhost:8080 に送る
ListenAndServe が受信 → ServeMux が /json にマッチ → handleJSON を呼ぶ
handleJSON → writeJSON が Content-Type ヘッダとJSONボディを書き出す
OS 経由でレスポンスがブラウザへ戻る
fetch の Response を await r.json() → JSオブジェクト化 → <pre> に表示
ここで見えるキー情報
ステータス: 200
レスポンスヘッダ: Content-Type: application/json; charset=utf-8
ボディ: {"message":"hi","time":"2025-10-15T...Z"}（書式は time.RFC3339）

7. なぜ Cookie で「同じ人」を判定できるの？
初回 /set-cookie のレスに Set-Cookie: sid=abc... を付ける
ブラウザは同一オリジンの次回以降のリクエストに自動で Cookie: sid=abc... を添付
サーバは r.Cookie("sid") で受け取り、**サーバ側ストレージ（ここでは Map）**のキーとして照合
だから「誰が誰か」を継続的に参照できる
本番は Map ではなく Redis/DB、Cookieは HttpOnly + Secure で守るのが基本

8. つまずいたら“可視化”してみる
ヘッダを全部見る：curl -i / Networkタブ
サーバ側で中身を出す：例）handleEcho の最初で io.ReadAll(r.Body) をログ出力
処理順を色付け：logging() で fmt.Println("[ENTER]", r.URL.Path)、ハンドラの最後で fmt.Println("[LEAVE]")

9. ここまでの“原理”のエッセンス
HTTPは「メソッド・パス・ヘッダ・ボディ」を送る／受けるだけ
**ServeMux（ディスパッチャ）**が「どのハンドラに渡すか」を決める
ハンドラがステータス・ヘッダ・ボディを作る（Content-Type はとくに大事）
Cookieは「小さな名札」。IDだけをクッキーに、本体はサーバ側に保存
fetch は結果を Promise で返し、JSでパースしてDOMに描画
この5点が腹落ちすれば、他言語・他フレームワークでも見通しが効くようになるよ。
さらに掘りたい箇所（例：Keep-Alive、HTTP/2、エラー設計、テンプレートSSRなど）があれば、該当コードに追記する形で“動く可視化”を足していこう🌙