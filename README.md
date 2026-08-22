# lumitree 🌲✨

TimeTree public calendar standardization bridge & lightweight proxy for CLI, iCalendar (.ics), and REST API.

---

## 🎯 What is lumitree?

`lumitree` は、TimeTree の公開カレンダー情報を取得・標準化し、**CLI**、**標準 iCalendar (.ics)**、および **OpenAPI 準拠の REST API** として提供する Go 製の軽量 Adapter / Proxy ツールです。

- **CSRF / セッション管理の完全隠蔽**: TimeTree 内部 API とのハンドシェイクを自動処理。
- **標準データ正規化**: ミリ秒タイムスタンプを ISO8601 / JST へ変換。
- **マルチインターフェース**:
  - 🖥️ **CLI**: `lumitree get <calendar_id>`, `lumitree ics <calendar_id>`
  - 🌐 **HTTP Server**: `lumitree serve --port 8080` (キャッシュ付き REST API & Webcal 配信)
- **軽量 & 安定稼働**: 自宅 k8s や最小 VPS でもメモリ数十 MB で常駐可能。

---

## 📚 ドキュメント

- 🗺️ **[開発ロードマップ (ROADMAP.md)](ROADMAP.md)**: 全フェーズの開発計画・テスト基準・品質ゲート
- 🏛️ **[システムアーキテクチャ & 全体像 (docs/architecture.md)](docs/architecture.md)**: Mermaid によるアーキテクチャ図・責務境界・データフロー

---

## 🚦 現在のステータス

- **Phase 0 (全体設計 & スキーマ策定)** 進行中 🛠️
