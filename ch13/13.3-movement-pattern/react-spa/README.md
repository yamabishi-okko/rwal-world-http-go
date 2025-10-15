### React SPA（Vite）について

cd react-spa<br>
npm in<br>
npm run dev<br>
# http://localhost:5173/<br>

React SPA＝最初から最後までブラウザで描く方式、SSR+Ajax＝初回はサーバがHTMLを作り、その後は必要データだけJSONでもらって一部更新だよ。<br>
### 一番の違い（どこでHTMLを作るか）<br>

#### React SPA（CSR）<br>
初回：薄いHTML＋大きめJSを受け取り、ブラウザ側で初回描画<br>
以降：JSON/APIを取りつつ、ブラウザ内で差分描画（仮想DOM）<br>

#### SSR + Ajax<br>
初回：サーバが完成HTMLを返す（SEO＆初速◎）<br>
以降：ボタン等で JSON(API) を取り、ページの一部だけJSで更新<br>

### 体験・性能の違い<br>
初回表示<br>
SPA：JSの読み込み完了まで白っぽいことがある<br>
SSR+Ajax：即コンテンツが出る（HTML完成品）<br>

### 以降の操作<br>
どちらもJSONでサクサク更新可能（SPAは標準、SSR+Ajaxは“足し算”）<br>

### SEO<br>
SPA：対策しないと弱い（プリレンダー/SSRが必要）<br>
SSR+Ajax：初回HTMLがあるので強い<br>

### JS依存度<br>
SPA：JS必須（切るとほぼ動かない）<br>
SSR+Ajax：初回はJSなしでも読める（Ajax部分はJS必要）<br>

### 開発・運用の違い<br>
実装の考え方<br>
SPA：UIロジックと状態管理をフロントに集中（ルーターもJS）<br>
SSR+Ajax：サーバのテンプレが主、足りない所だけAPI＋JSで補強<br>

### 複雑<br>
SPA：フロント設計（状態/ルーティング/ビルド）がコア<br>
SSR+Ajax：サーバ・テンプレ＋小さめJSで始めやすいが、API設計も必要<br>

### キャッシュ<br>
SPA：初回JSをCDNで効かせやすい／データはAPIでSWR等<br>
SSR+Ajax：HTMLキャッシュ＋APIキャッシュの二段構えがしやすい<br>

### DevToolsでの見分け方（Network）<br>
SPA：初回に document + 大きめJS、以降はfetch/XHR(JSON) と画像など。document はほぼ出ない<br>
SSR+Ajax：初回 document(HTML) で中身が見える → ボタン操作で /api/... のJSON が追加<br>
使い分けの目安<br>
SPAが向く：複雑なUI/フォーム、画面内インタラクションが多い業務アプリ、オフライン・PWA<br>
SSR+Ajaxが向く：コンテンツ中心・SEO重視・まずは読みやすく早く出したいサイトに少しの動きを足す<br>

### 超ミニ例の発想<br>
SPA：<br>
GET /index.html（薄い）→ 2) app.jsが初回描画 → 3) GET /api/items でJSON→描画<br>
SSR+Ajax：<br>
GET / → 完成HTML（リスト入り）<br>
「更新」クリック → GET /api/items のJSONだけ取り直し→一部だけ置換<br>