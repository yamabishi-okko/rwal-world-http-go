### サーバーサイドレンダリング+Ajaxについて

起動方法<br>
cd rwal-world-http-go/go-ssr-ajax<br>
go mod init example.com/go-ssr-ajax   # モジュール名は何でもOK<br>
go run .<br>
# => Go SSR at :8082  と出れば成功<br>


ざっくり違い<br>

#### サーバーサイドレンダリング（SSR）だけ<br>
初回も以降も：画面更新＝サーバが毎回 完成HTML を返す<br>
画面はフルリロード（document リクエストが毎回走る）<br>
View Source に中身（時刻や一覧）が常に埋め込まれている<br>
JSをオフにしてもほぼ成立（リンクで遷移・再読込すれば更新）<br>

#### SSR + Ajax<br>
初回：SSRと同じで完成HTMLを返す（速い・SEO◎）<br>
その後の更新：ボタンなどの操作で JSONだけをAPIから取得 → ページの一部だけ書き換え<br>
画面はフルリロードしない（Networkに XHR/fetch が並ぶ、document は増えない）<br>
JSが必要（Ajax部分はJSでDOM更新）<br>

### 何が嬉しい？

#### SSRだけ<br>
実装がシンプル、SEOも強い、JS不要でも閲覧可<br>
ただし操作のたびに全体を描き直すので体験が重くなりがち<br>

#### SSR+Ajax
初回の速さ・SEOはSSRのまま<br>
以降の操作は必要データだけを取りに行き部分更新、体験が軽い<br>
ただしAPI + フロントJSの実装分、SSRだけより作業は増える<br>

### DevToolsでの見分け方（実践）
SSRだけ：遷移や更新のたびに Type = document（HTML）が走る。/api/* の JSONは出てこない。<br>
SSR+Ajax：初回は document、その後の操作では /api/...（Type = fetch/XHR, JSON） が増え、documentは増えない。<br>

### コードの差（概念・最小）
SSRだけ：<br>
ルート / でテンプレ＋DB → HTML返す。更新はリンク/再読込のみ。<br>
SSR+Ajax：<br>
ルート / はSSRでHTML返す<br>
追加で /api/state を作る（JSON返す）<br>
フロントで fetch('/api/state') → 受け取ったJSONで innerHTML 等を差し替え<br>
つまり「API（JSON）を足して、JSで部分更新する」のが“SSR+Ajax”。<br>
迷ったらまずSSRで作り、重い部分だけAjax化が入門に最適だよ。🌙<br>