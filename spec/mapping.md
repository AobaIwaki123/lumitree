# TimeTree 生データ ↔ lumitree 標準モデル マッピング仕様書

本ドキュメントでは、TimeTree 内部 Web API の生レスポンスと、`lumitree` が提供する標準モデル（OpenAPI / `pkg/model`）との対応関係・変換ルールを定義します。

---

## 1. カレンダー基本情報 (Calendar)

### アップストリーム: `GET https://timetreeapp.com/api/v2/public_calendars/{aliasCode}`
```json
{
  "public_calendar": {
    "id": 46438,
    "alias_code": "ilife_official",
    "name": "iLiFE!",
    "overview": "iLiFE!のスケジュールです。",
    "images": {
      "cover": {
        "url": "https://attachments.timetreeapp.com/...",
        "thumbnail_url": "https://attachments.timetreeapp.com/..."
      }
    },
    "links": {
      "twitter": "https://x.com/iLiFE_official",
      "instagram": null
    }
  }
}
```

### マッピング定義
| TimeTree 生フィールド | lumitree 標準フィールド | 型 | 変換・正規化ルール |
| :--- | :--- | :--- | :--- |
| `public_calendar.id` | `id` | `string` | 文字列化（例: `"46438"`） |
| `public_calendar.alias_code` | `aliasCode` | `string` | そのまま保持 |
| `public_calendar.name` | `title` | `string` | 空文字の場合は `aliasCode` をフォールバック |
| `public_calendar.overview` | `description` | `string` | `null` の場合は空文字 `""` に置換 |
| `public_calendar.images.cover.url` | `coverImageUrl` | `string` (URI) | `null` または空の場合は未設定 |
| `public_calendar.links` | `snsLinks` | `object` | `twitter`, `instagram`, `website` を抽出（`null` は除外） |

---

## 2. イベント情報 (Event)

### アップストリーム: `GET https://timetreeapp.com/api/v2/public_calendars/{aliasCode}/public_events`
```json
{
  "paging": {
    "current_page": 1,
    "total_pages": 1,
    "total_count": 12
  },
  "public_events": [
    {
      "id": 12345678,
      "uuid": "98765432-abcd-ef01-2345-6789abcdef01",
      "title": "MEGALiFE! 先行物販＠Kアリーナ横浜",
      "description": "物販に関するご案内...",
      "start_at": 1787616000000,
      "end_at": 1787616000000,
      "all_day": false,
      "start_timezone": "Asia/Tokyo",
      "end_timezone": "Asia/Tokyo",
      "location": "Kアリーナ横浜",
      "url": "https://...",
      "images": [
        {
          "url": "https://..."
        }
      ]
    }
  ]
}
```

### マッピング定義
| TimeTree 生フィールド | lumitree 標準フィールド | 型 | 変換・正規化ルール |
| :--- | :--- | :--- | :--- |
| `id` | `id` | `string` | 文字列化（例: `"12345678"`） |
| `uuid` | `uuid` | `string` | そのまま保持 |
| `title` | `title` | `string` | 改行コードを除去しトリム |
| `description` | `description` | `string` | `null` の場合は空文字 `""` |
| `start_at` | `startAt` | `string` (ISO8601) | ミリ秒タイムスタンプ (`1787616000000`) を `time.UnixMilli` でパースし、`start_timezone`（デフォルト: `Asia/Tokyo`）の ISO8601 文字列（`2026-08-25T09:00:00+09:00`）に変換 |
| `end_at` | `endAt` | `string` (ISO8601) | `start_at` と同様に変換。`end_at <= start_at` の場合は `start_at + 1時間` または同値として処理 |
| `all_day` | `allDay` | `boolean` | `true` の場合、iCal 出力時に `VALUE=DATE` 形式とする |
| `start_timezone` | `timezone` | `string` | `null` または空の場合は `"Asia/Tokyo"` |
| `location` | `location` | `string?` | `null` または空文字の場合は `null` |
| `url` | `url` | `string?` | `null` または空文字の場合は `null` |
| `images[].url` | `imageUrls` | `string[]` | 画像オブジェクト配列から `url` 文字列の配列を抽出 |

---

## 3. iCalendar (RFC 5545) へのマッピング

`lumitree ics` コマンドおよび `GET /events.ics` での変換仕様です。

| iCalendar プロパティ | 設定値・ルール |
| :--- | :--- |
| `PRODID` | `-//lumitree//lumitree 1.0.0//EN` |
| `X-WR-CALNAME` | カレンダーの `title`（例: `iLiFE!`） |
| `X-WR-TIMEZONE` | `Asia/Tokyo` |
| `UID` | `{calendarId}-{eventId}@lumitree`（永続的・決定論的UID） |
| `SUMMARY` | イベントの `title` |
| `DESCRIPTION` | イベントの `description`（改行は `\n` でエスケープ） |
| `LOCATION` | イベントの `location` |
| `URL` | イベントの `url` |
| `DTSTART` | 終日: `20260825` / 時間指定: `TZID=Asia/Tokyo:20260825T090000` |
| `DTEND` | 終日: `20260826` (翌日) / 時間指定: `TZID=Asia/Tokyo:20260825T120000` |
| `LAST-MODIFIED` | イベント生成日時または現在日時 |
