# study-clean-architecture

ブログシリーズ「[Go言語で実装して理解するクリーンアーキテクチャ](https://log4me.dev/blog/clean-architecture-chapter1/)」のサンプルコードです。

タスク管理のREST APIを題材に、1つのファイルにすべてを書いた実装（`cmd/naive-api`）から始めて、クリーンアーキテクチャの層（Entities・Use Cases・Interface Adapters・Frameworks & Drivers）に分けて作り直していきます。

## 必要なもの

- Go 1.25 以降
- Docker（PostgreSQLの起動と、統合テスト・E2Eテストの実行に使う）

## 動かし方

```bash
# PostgreSQLを起動する（db/schema.sql でテーブルが作られる）
docker compose up -d

# APIを起動する
export DATABASE_URL='postgres://app@localhost:5432/tasks?sslmode=disable'
export API_TOKENS='my-local-token:alice'   # 「トークン:ユーザーID」をカンマ区切りで指定する（自分で決めた値にする）
go run ./cmd/api

# 別のターミナルから操作する
curl -X POST localhost:8080/tasks -H 'Authorization: Bearer my-local-token' -d '{"title":"牛乳を買う"}'
curl localhost:8080/tasks -H 'Authorization: Bearer my-local-token'
curl -X POST localhost:8080/tasks/<タスクID>/complete -H 'Authorization: Bearer my-local-token'
```

PostgreSQLを使わずに、メモリ上に保存して動かすこともできます（プロセスを終了するとデータは消えます）。

```bash
STORAGE=memory API_TOKENS='my-local-token:alice' go run ./cmd/api
```

同じユースケースを、コマンドラインから使うこともできます。

```bash
go run ./cmd/taskctl -user alice add '牛乳を買う'
go run ./cmd/taskctl -user alice list
go run ./cmd/taskctl -user alice done <タスクID>
```

### 環境変数

| 変数 | 説明 | 既定値 |
|------|------|--------|
| `HTTP_ADDR` | HTTPサーバーが待ち受けるアドレス | `:8080` |
| `STORAGE` | 保存先（`postgres` または `memory`） | `postgres` |
| `DATABASE_URL` | PostgreSQLの接続URL（`STORAGE=postgres` のとき必須） | なし |
| `API_TOKENS` | APIトークンとユーザーIDの対応（必須） | なし |
| `LOG_LEVEL` | ログのレベル（`debug` / `info` / `warn` / `error`） | `info` |
| `NOTIFY_WEBHOOK_URL` | タスクの完了を知らせるWebhookのURL | なし（通知しない） |

トークンなどの秘密の値は、コードやリポジトリに書かず、環境変数で渡してください（`.env` ファイルは `.gitignore` の対象にしています）。

## テスト

```bash
# ユニットテスト（Docker不要）
go test -race ./...

# PostgreSQLを使う統合テストとE2Eテスト（Dockerが必要）
go test -race -tags=integration ./...
```

`internal/archtest` のテストは、依存性のルール（内側の層が外側の層をimportしていないこと）を検査します。

## コードの生成

```bash
# sqlc（SQL → Goのコード）と mockgen（インターフェース → モック）
go generate ./...
```

## ディレクトリ構成

```
cmd/
  api/               HTTPのAPIのエントリーポイント（Composition Root）
  taskctl/           CLIのエントリーポイント
  naive-api/         第3章の「アーキテクチャなし」の実装（比較用）
internal/
  domain/            Entities：エンティティ・値オブジェクト・ドメインのエラー
  usecase/           Use Cases：ユースケースと、外の世界に求めるインターフェース（ポート）
  adapter/           Interface Adapters
    controller/        HTTPリクエスト → ユースケースの入力
    presenter/         ユースケースの出力 → HTTPレスポンス
    middleware/        認証・リクエストID・アクセスログ
    decorator/         ユースケースにログ出力を付け足すデコレータ
    cli/               コマンドライン引数 → ユースケースの入力
    gateway/           リポジトリや通知の実装（postgres / memory / notifier）
  infrastructure/    Frameworks & Drivers：設定・ログ・DB接続・ルーター・HTTPサーバー
  archtest/          依存性のルールを検査するテスト
  testutil/          テスト用のヘルパー（PostgreSQLのコンテナの起動など）
db/                  スキーマとsqlc用のクエリ
test/e2e/            アプリケーションを起動して外側から操作するE2Eテスト
```

## 記事との対応

| 章 | 内容 |
|----|------|
| [第3章](https://log4me.dev/blog/clean-architecture-chapter3/) | アーキテクチャなしで作る（`cmd/naive-api`） |
| [第4章](https://log4me.dev/blog/clean-architecture-chapter4/) | Entities（`internal/domain`） |
| [第5章](https://log4me.dev/blog/clean-architecture-chapter5/) | Use Cases（`internal/usecase`） |
| [第6章](https://log4me.dev/blog/clean-architecture-chapter6/) | Interface Adapters（`internal/adapter`） |
| [第7章](https://log4me.dev/blog/clean-architecture-chapter7/) | Frameworks & Drivers（`internal/infrastructure`、PostgreSQL） |
| [第8章](https://log4me.dev/blog/clean-architecture-chapter8/) | 依存性の注入と、グレースフルシャットダウン |
| [第9章](https://log4me.dev/blog/clean-architecture-chapter9/) | トランザクション |
| [第10章](https://log4me.dev/blog/clean-architecture-chapter10/) | 認証・エラーハンドリング・ログ |
| [第11章](https://log4me.dev/blog/clean-architecture-chapter11/) | テスト戦略（gomock・testcontainers・E2E） |
| [第12章](https://log4me.dev/blog/clean-architecture-chapter12/) | パッケージ構成と、依存性のルールの検査 |
| [第13章](https://log4me.dev/blog/clean-architecture-chapter13/) | 保存先の切り替え・CLI・通知の追加 |
