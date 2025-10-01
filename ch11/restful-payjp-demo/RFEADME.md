# RESTful PAY.JP Demo (Go)

このプロジェクトは、Go言語を使って **RESTful API にアクセスする流れ** を学ぶためのサンプルです。  
本来は教科書にある [PAY.JP](https://pay.jp/) API を題材にしていますが、ここでは安全のため **ダミーAPI [JSONPlaceholder](https://jsonplaceholder.typicode.com/)** を利用しています。

---

## 📦 セットアップ

### 1. プロジェクト作成
```bash
mkdir restful-payjp-demo
cd restful-payjp-demo
go mod init example.com/restful-payjp-demo
```

▶ 実行方法
go run main.go

実行例
== 顧客一覧（ダミーAPI） ==
ID=1 Title=sunt aut facere repellat provident occaecati excepturi optio reprehenderit
ID=2 Title=qui est esse
ID=3 Title=ea molestias quasi exercitationem repellat qui ipsa sit aut
...

💡 仕組み
本来の https://api.pay.jp/v1/customers の代わりに
https://jsonplaceholder.typicode.com/posts?userId=1 を叩いています。
Authorization: Bearer ... ヘッダをつける部分は 本物のAPIを想定した練習 です。
返ってきた JSON を Go の構造体にマッピングし、一覧表示しています。

⚠ 注意点
このサンプルは ダミーAPI用 です。データは実際には保存されません。
本当に PAY.JP を使う場合は公式のテストキーを発行してください。
APIキーは 絶対にコードに直書きせず、環境変数などで管理してください。