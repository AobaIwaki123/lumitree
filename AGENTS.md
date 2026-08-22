# lumitree 開発・協業規約 (AGENTS.md)

本ドキュメントは、AIエージェントおよび開発者が `lumitree` リポジトリで作業する際の共通規約を定義します。

---

## 0. リポジトリ概要

### プロジェクト情報

| 項目 | 内容 |
| :--- | :--- |
| リポジトリ | `github.com/AobaIwaki123/lumitree` |
| 言語 | Go 1.23 |
| ライセンス | MIT |
| コンテナレジストリ | `ghcr.io/aobaiwaki123/lumitree` |

**概要**: TimeTree の公開カレンダー情報を取得・標準化し、CLI / iCalendar (.ics) / OpenAPI 準拠 REST API として提供する Go 製の軽量 Adapter / Proxy ツール。

### ディレクトリ構成

```
lumitree/
├── api/                  # OpenAPI 仕様 (openapi.yaml) と oapi-codegen 設定
├── cmd/lumitree/         # エントリポイント (main.go)
├── pkg/
│   ├── api/              # oapi-codegen 自動生成クライアント (lumitree.gen.go)
│   ├── cache/            # In-Memory TTL キャッシュ
│   ├── config/           # 設定値 (環境変数・CLI フラグ)
│   ├── exporter/         # iCalendar (.ics) エクスポーター
│   ├── logger/           # 構造化ロガー
│   ├── model/            # 共通データモデル
│   ├── server/           # HTTP ハンドラ / REST API サーバー
│   └── timetree/         # TimeTree 内部 API クライアント (CSRF/セッション処理)
├── k8s/
│   ├── manifests/        # Kubernetes マニフェスト (kustomize)
│   └── argocd/           # ArgoCD Application 定義
├── docs/                 # 設計ドキュメント (architecture.md)
├── spec/                 # API マッピング仕様 (mapping.md)
├── scripts/              # 開発・運用補助スクリプト
├── testdata/             # テスト用フィクスチャ
└── tools/                # 開発ツール定義
```

### 主要スクリプト

| スクリプト | 用途 |
| :--- | :--- |
| `./scripts/worktree.sh create <branch>` | Git Worktree を作成してタスク作業領域を準備する |
| `./scripts/verify-all.sh` | lint / test / Schema Drift Check を一括実行（Push 前に必ず実行） |
| `./scripts/verify-deploy.sh` | Kubernetes 上の実稼働状態（ArgoCD / Pod / API 疎通）を一括検証する |
| `./scripts/create-release-pr.sh` | リリース PR を手動で起票する |
| `./scripts/sync-release-pr.sh` | リリース PR に最新の変更を同期する |
| `./scripts/update-deps.sh` | Go モジュール依存関係を更新する |

### CI/CD ワークフロー

| ワークフロー | トリガー | 内容 |
| :--- | :--- | :--- |
| `ci.yml` | PR / push | `golangci-lint` + `go test -race` + Schema Drift Check |
| `docker-publish.yml` | `release` ブランチへの push | コンテナイメージのビルド・GHCR Push |
| `release-pr.yml` | `main` ブランチへの push | リリース用ステージングブランチと PR を自動生成 |
| `tag-on-release-merge.yml` | `release` ブランチへのマージ | Git タグ発行・GoReleaser によるバイナリ配布 |
| `release.yml` | タグ push | GoReleaser によるリリース成果物の公開 |
| `live-monitor.yml` | スケジュール | 本番 API の死活監視 |

### コード生成

API クライアントは Schema-First で管理されており、手動編集禁止のファイルが存在します。

```bash
# api/openapi.yaml を変更したら必ず再生成する
go generate ./...
```

生成対象: `pkg/api/lumitree.gen.go`（`oapi-codegen` による自動生成）

---

## 1. Git & Pull Request & Worktree 運用ルール

- **`main` ブランチへの直接コミットおよび直接 Push は禁止**します。
- **すべてのタスク作業は Git Worktree（`.worktrees/<branch-name>`）を作成して行います**。ルート作業領域（`main` 等）での直接作業は禁止します（`./scripts/worktree.sh create <branch-name>` を活用）。
- **PR の単一責務の原則 (Single Responsibility PR)**: 各 PR は単一の明確な目的（バグ修正、機能追加、テスト追加、ドキュメント更新、CI/CD設定等）に限定し、無関係な変更の混在を禁止します。独立した改善は別ブランチを切って別 PR として起票します。
- **PR タイトルは、リリースノートだけを見て何をやったかが一目で分かるように、具体的かつ明瞭な日本語で記述します**（例: `feat: release ブランチ連動の Kubernetes 自動デプロイと検証スクリプトの追加`, `fix: .gitignore の修正による cmd/lumitree のコミット漏れ解消`）。
- 複数ステップの開発を行う場合は、親ブランチからの Stacked PR（積み上げ型 PR）として作成します。
- **エージェントによる PR の自律的なマージ・クローズは禁止**します。PR 作成と CI（Lint / Test）の通過確認までを作業範囲とし、マージはユーザーのレビュー・判断に委ねます。

---

## 2. ドキュメント & Mermaid 規約

- README、ROADMAP、設計仕様書などの公式ドキュメントでは、**原則として絵文字（emoji）を使用しません**。
- 清潔でプロフェッショナルな Markdown 記述を徹底します。
- **Mermaid 構文エラーの撲滅 (Double-Check 原則)**:
  - Mermaid 図を作成・更新する際は、ノードラベル、エッジテキスト（`-->|"..."|`）、SequenceDiagram の participant 名・Note 本文に含まれる特殊文字（`{`, `}`, `(`, `)`, `#`, `/`, `.`, `-` 等）を **必ずダブルクォート (`"..."`) でエスケープ・クォート** します。
  - 作成・編集後は、GitHub Markdown レンダラーで構文エラー（Parse error）が発生しないか **必ずセルフレビュー（Double-Check）** を実施してから PR を起票します。

---

## 3. 品質基準 (Quality Gate) & CI-Green 原則

- **CI が通るまで絶対にマージしない**:
  - すべての PR は、GitHub Actions CI（`golangci-lint`、`go test -race`、`Schema Drift Check` 等）が 100% PASS (Green) することを確認するまで、**絶対にマージを行ってはなりません**。
  - CI が `pending`（実行中）または `failure`（失敗）の状態でのマージは例外なく禁止します。
- コードコメント（パッケージコメント、エクスポート型/関数のコメント）を適切に付与し、Linter 警告を 0 に保ちます。
- コミット・Push 前に必ず `./scripts/verify-all.sh` をローカルで実行し、事前検証を徹底します。
