### サーバーサイドレンダリングについて

起動方法<br>
cd rwal-world-http-go/go-ssr<br>
go mod init example.com/go-ssr   # モジュール名は何でもOK<br>
go run .<br>
# => Go SSR at :8081  と出れば成功<br>

1) DevTools の Network タブ（必須）

ページ開く → DevTools → Network

右クリックして Disable cache をON、Preserve log もON

⌘R（ハードリロード推奨）

見るポイント：

Document の1本太いリクエスト（/）が毎回走る

Type = document、Content-Type = text/html（JSONやXHRは出ない）

Waterfall がほぼ document だけ（JSバンドルは無い/最小）

TTFB と Size (transferred)：遷移・更新のたびにHTMLを丸ごと受け取ってるのが見える

比較メモ：SPA だと初回に document + 大きめJS、以降は fetch/XHR (JSON) が並ぶようになる。
SSR+Ajax だと初回は document、その後は /api/state (JSON) が出現。

2) View Source vs Elements（DOM差）

右クリック → View Page Source：
HTMLの中に値（Now, Items）が埋め込まれているのが見える（= サーバでテンプレ化済み）。

Elements（検証）→ そのまま同じ内容。
→ SSRは “届いたHTMLそのまま” を描画。JSでの差分復元は無し（=ハイドレーションなし）。

比較メモ：SPAは View Source が “空に近いシェル”＋スクリプト参照だけで、実データは載ってない。
Next(SSR+SPA)は View Sourceに初期値あり → 直後にクライアントJSがハイドレート。

3) リロード挙動（全描画のやり直し）

画面の Reloadリンク か ⌘R を何度か。

Network に 毎回 document が記録、画面はフルリロードされる。

“時刻”は必ず変わる（=サーバで毎回レンダリング）。

比較メモ：SSR+Ajaxは「ボタン押下」では documentは増えず、/api/state (JSON) だけ増える。
SPAは「画面遷移」でも document は出ず、JS内の仮想ルーターで描画が切り替わる。

4) JS依存度テスト（オフにしてみる）

DevTools → Command Palette（⌘⇧P）→ “Disable JavaScript”

ページ更新：SSRはそのまま表示される（HTML完成品だから）。
→ JSを切っても最低限成立がSSRの強み。

比較メモ：SPAはJSを切るとほぼ何も動かない。
Next(SSR+SPA)は初期表示は出るが、その後の操作はJS無しだと非活性。

5) curl でヘッダ確認（HTTP目線）
curl -i http://localhost:8081/


見るポイント：

HTTP/1.1 200 OK

Content-Type: text/html; charset=utf-8

ボディが完成HTML（Now/Itemsが文字として入ってる）

比較メモ：AjaxのAPIは Content-Type: application/json。
SPA側のAPI呼び出しも JSON を返すのが基本。

ぱっと見の“違いサイン”まとめ

SSR：毎回 document(HTML)。View Sourceに中身あり。JS不要でもOK。

SSR+Ajax：初回 document、その後は JSON（/api/*）が増える。

SPA：初回に JS束、以降は JSON中心。View Sourceは空っぽ気味。

SPA+SSR：初回 documentに初期値あり→直後にHydration、以降 JSON中心。

次は go-ssr-ajax を起動して、同じチェックをやろう。
Networkに /api/state（JSON）が現れるはず。そこまでいけば、4方式の見分けは完璧だよ🌙