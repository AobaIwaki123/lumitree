# lumitree システムアーキテクチャ & 全体像

`lumitree` は、TimeTree の公開カレンダー情報を取得・標準化し、CLI、標準 iCalendar (.ics)、および OpenAPI 準拠の REST API として提供する軽量な Adapter / Bridge ツールです。

---

## 1. 全体アーキテクチャ図 (System Overview)

```mermaid
flowchart TD
    subgraph External["外部データソース"]
        TT["TimeTree Public Calendar<br/>(非公式内部API / Web)"]
    end

    subgraph LumitreeApp["lumitree (Go実装 / シングルバイナリ / コンテナ)"]
        Client["TimeTree Client<br/>(CSRF・Cookie・Header管理)"]
        Cache["In-Memory TTL Cache<br/>(過負荷・BAN防止)"]
        Normalizer["Data Normalizer<br/>(ISO8601 / JST変換)"]
        ICS["iCalendar Exporter<br/>(RFC 5545 準拠 .ics 生成)"]
        CLI["CLI Runner<br/>(get / ics / serve)"]
        HTTPServer["HTTP Server<br/>(REST API / iCal配信 / healthz)"]

        Client <--> Cache
        Client --> Normalizer
        Normalizer --> ICS
        Normalizer --> CLI
        ICS --> CLI
        Normalizer --> HTTPServer
        ICS --> HTTPServer
    end

    subgraph Downstream["下流アプリケーション & 自宅インフラ"]
        K8s["自宅 k8s クラスタ<br/>(ArgoCD / Ingress)"]
        Discord["Discord / LINE Bot<br/>(イベント通知)"]
        GCal["Google / Apple カレンダー<br/>(Webcal .ics 購読)"]
        Portal["自作イベントポータル<br/>(REST API 連携)"]
    end

    TT <-->|HTTPS / CSRF| Client
    K8s -.->|ホスト & 運用| LumitreeApp
    HTTPServer -->|REST JSON| Discord
    HTTPServer -->|REST JSON| Portal
    HTTPServer -->|webcal .ics| GCal
    CLI -->|JSON / Pipe| Discord
```

---

## 2. 責務の分離と境界線 (Separation of Concerns)

| レイヤー | 責務（やること） | 責務外（やらないこと） |
| :--- | :--- | :--- |
| **lumitree (本プロジェクト)** | ・TimeTree 内部APIとのハンドシェイク・通信<br/>・データ正規化（ISO8601/JSTへの統一）<br/>・標準 iCal (`.ics`) / JSON 出力<br/>・OpenAPI 準拠のキャッシュ付き HTTP Proxy<br/>・自宅 k8s 用コンテナ提供 | ・Discord / LINE の通知・Bot ロジック<br/>・Google Calendar OAuth 認証<br/>・ポータルサイト用 DB 保存や UI 実装 |
| **下流アプリ (Bot / Portal)** | ・ユーザー向け通知、UI描画、独自DB管理 | ・TimeTree 固有の通信・CSRF解析 |
| **カレンダーアプリ (GCal/Apple)** | ・`.ics` URL を定期フェッチして予定表示 | ・自前でのデータスクレイピング |

---

## 3. データフロー

### ① CLI スタンドアロン実行時
```
[User / Cron] 
      │
      ▼ (lumitree ics ilife_official)
[CLI] ──► [Client] ──► [TimeTree HTML (CSRF)] ──► [TimeTree API]
                            │
                            ▼ (Raw JSON)
                      [Normalizer] ──► [ICS Exporter] ──► [Stdout / File (.ics)]
```

### ② HTTP Server (自宅 k8s デプロイ時)
```
[Google Calendar / Discord Bot / Web Portal]
      │
      ▼ (GET /api/v1/calendars/ilife_official/events.ics)
[HTTP Server (lumitree serve)]
      │
   [Cache Check] ──(Hit)──► [Return Cached .ics (Fast)]
      │ (Miss)
      ▼
[Client (Handshake & Fetch)] ──► [TimeTree]
      │
      ▼ (Normalize & Generate)
[Save to Cache] ──► [Return 200 OK + Content-Type: text/calendar]
```
