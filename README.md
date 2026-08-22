# lumitree

TimeTree public calendar standardization bridge & lightweight proxy for CLI, iCalendar (.ics), and REST API.

[![CI - Go](https://github.com/AobaIwaki123/lumitree/actions/workflows/ci-go.yml/badge.svg)](https://github.com/AobaIwaki123/lumitree/actions/workflows/ci-go.yml)
[![CI - Shell Scripts](https://github.com/AobaIwaki123/lumitree/actions/workflows/ci-scripts.yml/badge.svg)](https://github.com/AobaIwaki123/lumitree/actions/workflows/ci-scripts.yml)
[![CI - Workflows](https://github.com/AobaIwaki123/lumitree/actions/workflows/ci-workflows.yml/badge.svg)](https://github.com/AobaIwaki123/lumitree/actions/workflows/ci-workflows.yml)
[![CI - Live API Monitoring](https://github.com/AobaIwaki123/lumitree/actions/workflows/ci-live-monitor.yml/badge.svg)](https://github.com/AobaIwaki123/lumitree/actions/workflows/ci-live-monitor.yml)
[![GitHub Release](https://img.shields.io/github/v/release/AobaIwaki123/lumitree)](https://github.com/AobaIwaki123/lumitree/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

---

## 概要

`lumitree` は、TimeTree の公開カレンダー情報を取得・標準化し、**CLI**、**標準 iCalendar (.ics)**、および **OpenAPI 準拠の REST API** として提供する Go 製の軽量 Adapter / Proxy ツールです。

- **CSRF / セッション管理の完全隠蔽**: TimeTree 内部 API とのハンドシェイク（Cookie/CSRF 抽出）を自動処理。
- **標準データ正規化**: ミリ秒タイムスタンプを ISO8601 / JST へ高精度に変換。
- **In-Memory TTL キャッシュ**: TimeTree への過度なリクエストや IP BAN を防止（デフォルト 10 分）。
- **Webcal (.ics) 配信**: Google カレンダーや Apple カレンダーから URL 指定で自動購読可能。
- **Kubernetes Ready**: Distroless ベースの超軽量マルチアーキテクチャ OCI コンテナイメージ（数十 MB 以下で常駐）。
- **Go OpenAPI 自動生成クライアント**: `oapi-codegen` による型安全な Go クライアントライブラリを内包。

---

## インストール

### 方法 1: GitHub Releases からバイナリをダウンロード (推奨)

[GitHub Releases](https://github.com/AobaIwaki123/lumitree/releases) より、お使いの OS / アーキテクチャに合わせた最新バイナリをダウンロードして展開してください。

```bash
# macOS (Apple Silicon) の例
curl -sL https://github.com/AobaIwaki123/lumitree/releases/latest/download/lumitree_Darwin_arm64.tar.gz | tar -xz
sudo mv lumitree /usr/local/bin/
```

### 方法 2: Go Install
```bash
go install github.com/AobaIwaki123/lumitree/cmd/lumitree@latest
```

### 方法 3: Docker (GitHub Container Registry)
```bash
docker run --rm -p 8080:8080 ghcr.io/aobaiwaki123/lumitree:latest
```

---

## CLI の使い方

### 1. カレンダーイベントの取得 (`get`)

```bash
# テーブル形式でターミナル表示 (デフォルト)
lumitree get ilife_official

# 年・月・ページを指定して取得
lumitree get ilife_official --year 2026 --month 8 --page 1

# 正規化された JSON で出力 (jq 連携・スクリプト連携)
lumitree get ilife_official --json | jq '.events[] | {title: .title, start: .start_at}'
```

### 2. iCalendar (.ics) ファイルのエクスポート (`ics`)

```bash
# 標準出力へ出力
lumitree ics ilife_official

# ファイルに直接保存 (-o または --output)
lumitree ics ilife_official -o ilife_calendar.ics
```

### 3. HTTP プロキシサーバーの起動 (`serve`)

```bash
# デフォルト設定 (ポート 8080) で起動
lumitree serve

# ポートやキャッシュ TTL を指定して起動
lumitree serve --port 8080 --host 0.0.0.0 --cache-ttl 15m --log-format json
```

---

## HTTP REST API & Webcal エンドポイント

| メソッド | パス | 説明 | レスポンス形式 |
| :--- | :--- | :--- | :--- |
| `GET` | `/healthz` | ヘルスチェック (Kubernetes Liveness / Readiness) | `application/json` |
| `GET` | `/api/v1/calendars/{id}` | カレンダー基本情報 (ID, タイトル, アイコン等) | `application/json` |
| `GET` | `/api/v1/calendars/{id}/events` | イベント一覧 (クエリ: `year`, `month`, `page`) | `application/json` |
| `GET` | `/api/v1/calendars/{id}/events.ics` | iCalendar 配信 (Google Cal / Apple Cal 購読用) | `text/calendar` |

---

## Google カレンダー / Apple カレンダーでの自動同期

1. `lumitree serve` を自宅 Kubernetes や VPS で起動（例: `https://lumitree.example.com`）。
2. Google カレンダーの左メニュー「他のカレンダー」の「＋」をクリックし、「**URL で追加**」を選択。
3. カレンダーの URL に以下を入力：
   ```
   https://lumitree.example.com/api/v1/calendars/ilife_official/events.ics
   ```
4. これにより、TimeTree 上の最新イベントが Google カレンダー上に自動同期されます。

---

## 環境変数設定

CLI 引数のほか、以下の環境変数による設定に対応しています：

| 環境変数 | デフォルト値 | 説明 |
| :--- | :--- | :--- |
| `LUMITREE_PORT` | `8080` | HTTP サーバーのリッスンポート |
| `LUMITREE_HOST` | `0.0.0.0` | HTTP サーバーのリッスンホスト |
| `LUMITREE_CACHE_TTL` | `10m` | TimeTree API レスポンスのキャッシュ保持期間 |
| `LUMITREE_LOG_LEVEL` | `info` | ログレベル (`debug`, `info`, `warn`, `error`) |
| `LUMITREE_LOG_FORMAT`| `json` | ログ出力形式 (`json`, `text`) |

---

## Kubernetes & GitOps (ArgoCD) デプロイ

### Kubernetes マニフェストの適用
```bash
kubectl apply -k k8s/manifests/
```

### ArgoCD による自動同期
```bash
kubectl apply -n argocd -f k8s/argocd/app.yml
```

### デプロイ実稼働検証スクリプト
```bash
./scripts/verify-deploy.sh
```

---

## 開発とスキーマコード生成

```bash
# 全テスト & リント & スキーマ整合性検証 (Schema Drift Check)
./scripts/verify-all.sh

# OpenAPI 3.0.3 仕様書 (api/openapi.yaml) からクライアントコードを再生成
go generate ./...
```

---

## ライセンス

[MIT License](LICENSE) © 2026 Aoba Iwaki
