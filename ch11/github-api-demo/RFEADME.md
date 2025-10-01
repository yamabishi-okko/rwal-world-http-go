# GitHub API Demo (Go)

このプロジェクトは、Go言語を使って **GitHub REST API** にアクセスするサンプルです。  
Personal Access Token (PAT) を利用して認証し、ユーザー情報やリポジトリ一覧を取得します。

---

## 📦 セットアップ

### 1. このリポジトリをクローン
```bash
git clone https://github.com/your-username/github-api-demo.git
cd github-api-demo
```

### 2. Goモジュールを初期化
```bash
go mod init example.com/github-api-demo
go mod tidy
```

### 3. GitHub Personal Access Token の準備
・GitHub の Settings → Developer settings → Personal access tokens から発行
・Scopes（権限）は最低限以下を推奨
・read:user（ユーザープロフィールの取得）
・user:email（メールアドレスの取得）
・repo（リポジトリの一覧取得）

発行されたトークンはコピーして環境変数に設定します。


▶ 実行方法
go run main.go


実行例
== Authenticated as: yamabishi-okko (name=yamabishi-okko)
== Emails:
- example@gmail.com primary=true verified=true
== Repos (latest 5):
- rwal-world-http-go (private=false)
- quiz-billionaire (private=false)
- Pentaroll-react (private=false)
- deka-fantasy (private=true)
- deka-fantasy-be (private=true)
== Rate limit headers:
X-RateLimit-Limit: 5000
X-RateLimit-Remaining: 4990

⚠ 注意点
トークンは忘れた場合 再発行が必要 です（再表示できません）
公開リポジトリに push すると自動で無効化される可能性があります
