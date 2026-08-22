# lumitree システムアーキテクチャ & 全体像

`lumitree` は、TimeTree の公開カレンダー情報を取得・標準化し、CLI、標準 iCalendar (.ics)、および OpenAPI 準拠の REST API として提供する軽量な Adapter / Bridge ツールです。

---

## 1. 全体アーキテクチャ図 (System Overview)

```mermaid
flowchart TB
    subgraph External ["外部サービス (データソース)"]
        TT["TimeTree Public Calendar<br/>(非公式内部API / Webフロントエンド)"]
    end

    subgraph LumitreeApp ["lumitree (Go実装 / シングルバイナリ / コンテナ)"]
        direction TB
        
        subgraph Core ["Core Domain & Adapter"]
            Client["TimeTree Client<br/>・CSRF Handshake<br/>・Session Cookie管理<br/>・X-TimeTreeA 付与"]
            Cache["In-Memory TTL Cache<br/>(レートリミット・過負荷防止)"]
            Normalizer["Data Normalizer<br/>・TimeTree生データ → 標準Model<br/>・ISO8601 / JSTタイムゾーン変換"]
        end

        subgraph Exporters ["Export & Formatting"]
            ICS["iCalendar Exporter<br/>(RFC 5545 準拠 .ics 生成)"]
            JSONExp["JSON / Table Formatter"]
        end

        subgraph Interfaces ["Interfaces (入出力口)"]
            CLI["CLI Runner (cobra/pflag)<br/>・lumitree get &lt;id&gt;<br/>・lumitree ics &lt;id&gt;<br/>・lumitree serve"]
            HTTPServer["HTTP Server (REST / iCal)<br/>・GET /api/v1/calendars/{id}<br/>・GET /api/v1/calendars/{id}/events<br/>・GET /api/v1/calendars/{id}/events.ics<br/>・GET /healthz"]
        end

        Client <--> Cache
        Client --> Normalizer
        Normalizer --> ICS
        Normalizer --> JSONExp
        
        JSONExp --> CLI
        ICS --> CLI
        
        Normalizer --> HTTPServer
        ICS --> HTTPServer
    end

    subgraph Downstream ["下流アプリケーション & インフラ"]
        K8s["自宅 Kubernetes (k8s) クラスタ<br/>(ArgoCD / Deployment / Service / Ingress)"]
        Discord["Discord Bot / LINE Bot<br/>(イベント通知・リマインド)"]
        GCal["Google カレンダー / Apple カレンダー<br/>(Webcal .ics URL購読)"]
        Portal["自作イベントポータル (Web UI)<br/>(REST API経由フェッチ)"]
    end

    TT <== "HTTPS (CSRF/Cookie)" ==> Client
    K8s -.->|ホスト & 運用| LumitreeApp
    HTTPServer -->|OpenAPI REST JSON| Discord
    HTTPServer -->|OpenAPI REST JSON| Portal
    HTTPServer -->|webcal:// .../events.ics| GCal
    CLI -->|Stdout (JSON / Pipe)| Discord
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
