# lumitree スタートアップガイド (Getting Started)

本ガイドでは、`lumitree` を使って TimeTree の公開カレンダーを **CLI で確認**、**Google / Apple カレンダーに自動同期**、および **REST API でプログラム連携** するまでの手順をステップ・バイ・ステップで解説します。

---

## 目次
1. [Step 1: TimeTree 公開カレンダーの ID を確認する](#step-1-timetree-公開カレンダーの-id-を確認する)
2. [Step 2: CLI をインストールして予定を確認する](#step-2-cli-をインストールして予定を確認する)
3. [Step 3: Google カレンダー / Apple カレンダーに自動同期する](#step-3-google-カレンダー--apple-カレンダーに自動同期する)
4. [Step 4: REST API から JSON データを取得する](#step-4-rest-api-から-json-データを取得する)
5. [Step 5: Kubernetes / Docker で常駐運用する](#step-5-kubernetes--docker-で常駐運用する)

---

## Step 1: TimeTree 公開カレンダーの ID を確認する

ブラウザで購読したい TimeTree の公開カレンダーページを開きます。

* 例: `https://timetreeapp.com/public_calendars/ilife_official`

URL の末尾にある文字列（上記の場合は **`ilife_official`**）がカレンダー ID（`calendarId` / `alias_code`）です。
数字のみの ID（例: `46438`）でも同様に使用できます。

---

## Step 2: CLI をインストールして予定を確認する

### 1. インストール
お使いの環境に合わせて以下のいずれかの方法でインストールします。

```bash
# macOS (Apple Silicon) の場合
curl -sL https://github.com/AobaIwaki123/lumitree/releases/latest/download/lumitree_Darwin_arm64.tar.gz | tar -xz
sudo mv lumitree /usr/local/bin/

# Go がインストールされている場合
go install github.com/AobaIwaki123/lumitree/cmd/lumitree@latest
```

### 2. ターミナルで予定を表示
```bash
lumitree get ilife_official
```

ターミナルに直近のイベント一覧が美しくテーブル表示されます。

### 3. iCalendar (.ics) ファイルを生成
```bash
lumitree ics ilife_official -o my_calendar.ics
```

カレントディレクトリに `my_calendar.ics` がエクスポートされます。

---

## Step 3: Google カレンダー / Apple カレンダーに自動同期する

TimeTree のイベントを定期的に自動同期（Webcal 購読）するには、`lumitree` の HTTP サーバー機能を使用します。

### 1. プロキシサーバーを起動
```bash
lumitree serve --port 8080
```

### 2. カレンダーアプリに登録
サーバーが稼働しているホスト（例: `https://lumitree.aooba.net` または `http://localhost:8080`）のエンドポイントを登録します：

* **購読 URL**:
  ```
  https://lumitree.aooba.net/api/v1/calendars/ilife_official/events.ics
  ```

#### 🍏 Apple カレンダー (iPhone / Mac)
1. 「ファイル」メニュー >「新規カレンダー照会...」を選択。
2. 上記の URL を入力して「照会」をクリック。
3. 自動更新頻度（例: 「毎日」または「毎時間」）を選択して保存。

#### 📅 Google カレンダー (Web)
1. 左サイドバーの「他のカレンダー」の横にある **「＋」** をクリック。
2. **「URL で追加」** を選択。
3. 上記の URL を入力して「カレンダーを追加」をクリック。
   > ※ Google の初回フェッチには数分〜数十分かかる場合があります。

---

## Step 4: REST API から JSON データを取得する

Discord Bot や LINE Bot、自作フロントエンド等で予定データを扱いたい場合、REST API を利用できます。

### カレンダー情報の取得
```bash
curl -s https://lumitree.aooba.net/api/v1/calendars/ilife_official | jq .
```

### イベント一覧の取得 (JSON)
```bash
curl -s "https://lumitree.aooba.net/api/v1/calendars/ilife_official/events?year=2026&month=8" | jq .
```

レスポンス例:
```json
{
  "calendar": {
    "id": "46438",
    "alias_code": "ilife_official",
    "title": "iLiFE!",
    "timezone": "Asia/Tokyo"
  },
  "events": [
    {
      "id": "10001",
      "title": "MEGALiFE! 先行物販＠Kアリーナ横浜",
      "start_at": "2026-08-25T09:00:00+09:00",
      "end_at": "2026-08-25T12:00:00+09:00",
      "all_day": false,
      "location": "Kアリーナ横浜"
    }
  ]
}
```

---

## Step 5: Kubernetes / Docker で常駐運用する

### Docker での即時起動
```bash
docker run -d --name lumitree -p 8080:8080 ghcr.io/aobaiwaki123/lumitree:latest
```

### Kubernetes (ArgoCD / GitOps) でのデプロイ
```bash
# マニフェストの直接適用
kubectl apply -k k8s/manifests/

# 稼働・疎通の自動検証
./scripts/verify-deploy.sh
```

---

## 📚 関連ドキュメント
* [システムアーキテクチャ & GitOps 運用設計 (architecture.md)](architecture.md)
* [OpenAPI 3.0.3 仕様書 (openapi.yaml)](../api/openapi.yaml)
