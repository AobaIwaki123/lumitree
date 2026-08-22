# lumitree 🌲✨

TimeTree public calendar standardization bridge & lightweight proxy for CLI, iCalendar (.ics), and REST API.

[![CI](https://github.com/AobaIwaki123/lumitree/actions/workflows/ci.yml/badge.svg)](https://github.com/AobaIwaki123/lumitree/actions/workflows/ci.yml)
[![Live API Monitoring](https://github.com/AobaIwaki123/lumitree/actions/workflows/live-monitor.yml/badge.svg)](https://github.com/AobaIwaki123/lumitree/actions/workflows/live-monitor.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

---

## 🎯 What is lumitree?

`lumitree` は、TimeTree の公開カレンダー情報を取得・標準化し、**CLI**、**標準 iCalendar (.ics)**、および **OpenAPI 準拠の REST API** として提供する Go 製の軽量 Adapter / Proxy ツールです。

- 🛡️ **CSRF / セッション管理の完全隠蔽**: TimeTree 内部 API とのハンドシェイクを自動処理。
- 🔄 **標準データ正規化**: ミリ秒タイムスタンプを ISO8601 / JST へ変換。
- ⚡ **In-Memory TTL キャッシュ**: TimeTree への過度なリクエストや IP BAN を防止。
- 📅 **Webcal (.ics) 配信**: Google カレンダーや Apple カレンダーから URL 指定で自動購読可能。
- ☸️ **Kubernetes Ready**: Distroless ベースの超軽量コンテナ（数十 MB 以下で常駐）。

---

## 🚀 クイックスタート (CLI)

### インストール
```bash
# リポジトリからビルド
git clone https://github.com/AobaIwaki123/lumitree.git
cd lumitree
go build -o bin/lumitree ./cmd/lumitree
```

### 使い方

#### 1. イベント一覧の取得 (テーブル表示)
```bash
./bin/lumitree get ilife_official
```

#### 2. 正規化された JSON の出力 (パイプ / jq 連携)
```bash
./bin/lumitree get ilife_official --json | jq .events[0]
```

#### 3. iCalendar (.ics) のエクスポート
```bash
# ファイルに出力
./bin/lumitree ics ilife_official --out ilife.ics
```

#### 4. HTTP プロキシサーバーの起動 (REST API / Webcal 配信)
```bash
./bin/lumitree serve --port 8080
```

---

## 🌐 HTTP API エンドポイント

| メソッド | パス | 説明 | レスポンス形式 |
| :--- | :--- | :--- | :--- |
| `GET` | `/healthz` | ヘルスチェック (Liveness / Readiness) | `application/json` |
| `GET` | `/api/v1/calendars/{id}` | カレンダー基本情報 | `application/json` |
| `GET` | `/api/v1/calendars/{id}/events` | イベント一覧 (クエリ: `year`, `month`, `page`) | `application/json` |
| `GET` | `/api/v1/calendars/{id}/events.ics` | iCalendar 配信 (Google Cal 購読用) | `text/calendar` |

---

## 📱 Google カレンダー / Apple カレンダーでの自動同期

1. `lumitree serve` を自宅 k8s や VPS で起動（例: `https://lumitree.example.com`）。
2. Google カレンダーの左メニュー「他のカレンダー」の「＋」をクリック → **「URL で追加」** を選択。
3. カレンダーの URL に以下を入力：
   ```
   https://lumitree.example.com/api/v1/calendars/ilife_official/events.ics
   ```
4. これだけで、TimeTree 上の最新イベントが Google カレンダー上に自動同期されます。

---

## 🐳 Docker & Kubernetes での起動

### Docker
```bash
docker build -t lumitree:latest .
docker run -d -p 8080:8080 --name lumitree lumitree:latest
```

### Kubernetes (Kustomize)
```bash
kubectl apply -k k8s/
```

---

## 📚 ドキュメント & 仕様

- 🏛️ **[システムアーキテクチャ (docs/architecture.md)](docs/architecture.md)**
- 🗺️ **[開発ロードマップ (ROADMAP.md)](ROADMAP.md)**
- 📑 **[OpenAPI 3.0.3 仕様書 (api/openapi.yaml)](api/openapi.yaml)**
- 📐 **[生データ ↔ 標準モデル マッピング仕様 (spec/mapping.md)](spec/mapping.md)**

---

## 📄 ライセンス

[MIT License](LICENSE) © 2026 Aoba Iwaki
