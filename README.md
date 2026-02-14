# Issues - AI駆動 Issue 管理システム

## 1. システム概要

Issues は、GitHub のような Issue 管理機能を提供するアプリケーションである。最大の特徴は、Issue が起票（open）されると Claude Code が自動的にタスクを実行し、完了時に通知を送る点にある。一人プロジェクト向けに設計されており、シンプルな認証モデルを採用する。

### 1.1 主な機能

- **プロジェクト管理**: プロジェクトを作成し、プロジェクト単位で Issue を管理
- **Issue ライフサイクル**: 起票 → AI 自動実行 → 完了報告
- **AI 自動実行**: Issue が open になると Claude Code がバックグラウンドでタスクを処理
- **通知**: Issue の完了・失敗をリアルタイムに通知（In-App + Webhook）
- **マルチプラットフォーム**: Web（React）+ モバイル（React Native）

### 1.2 全体アーキテクチャ

- **Backend**: Go（chi router, pgx/sqlx, JWT 認証）
- **Frontend Web**: React + TypeScript + Vite + Tailwind CSS
- **Frontend Mobile**: React Native + Expo
- **Database**: PostgreSQL
- **AI Runtime**: Claude Code CLI（サブプロセス実行）
- **認証**: Google / GitHub OAuth

## 2. データフロー設計

### 2.1 Issue 起票から AI 実行まで

1. ユーザーが Issue を作成（`POST /api/v1/projects/:pid/issues`）
2. Issue が `open` ステータスで保存される
3. `ai_jobs` テーブルに `pending` ジョブが挿入される
4. ワーカープール（goroutine）が `FOR UPDATE SKIP LOCKED` でジョブを取得
5. Claude Code CLI をサブプロセスとして実行（`--print` フラグ、タイムアウト30分）
6. 完了時: Issue を `completed` に更新、通知を作成
7. 失敗時: リトライ（最大3回）、全失敗で `closed` に更新、失敗通知を作成

### 2.2 認証フロー

1. ユーザーが Google / GitHub ログインを選択
2. OAuth プロバイダにリダイレクト
3. コールバックでトークン交換、ユーザー Upsert
4. JWT アクセストークン（15分）+ リフレッシュトークン（7日）を発行

## 3. API 設計

### 3.1 エンドポイント一覧

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/api/v1/auth/google` | - | Google OAuth リダイレクト |
| GET | `/api/v1/auth/google/callback` | - | Google コールバック |
| GET | `/api/v1/auth/github` | - | GitHub OAuth リダイレクト |
| GET | `/api/v1/auth/github/callback` | - | GitHub コールバック |
| POST | `/api/v1/auth/refresh` | - | トークンリフレッシュ |
| GET | `/api/v1/auth/me` | Bearer | 認証ユーザー情報 |
| GET | `/api/v1/projects` | Bearer | プロジェクト一覧 |
| POST | `/api/v1/projects` | Bearer | プロジェクト作成 |
| GET | `/api/v1/projects/:id` | Bearer | プロジェクト詳細 |
| PATCH | `/api/v1/projects/:id` | Bearer | プロジェクト更新 |
| DELETE | `/api/v1/projects/:id` | Bearer | プロジェクト削除 |
| GET | `/api/v1/projects/:pid/issues` | Bearer | Issue 一覧 |
| POST | `/api/v1/projects/:pid/issues` | Bearer | Issue 作成（→ AI 実行） |
| GET | `/api/v1/projects/:pid/issues/:id` | Bearer | Issue 詳細 |
| PATCH | `/api/v1/projects/:pid/issues/:id` | Bearer | Issue 更新 |
| POST | `/api/v1/projects/:pid/issues/:id/close` | Bearer | Issue クローズ |
| POST | `/api/v1/projects/:pid/issues/:id/reopen` | Bearer | Issue 再開 |
| GET | `/api/v1/notifications` | Bearer | 通知一覧 |
| POST | `/api/v1/notifications/:id/read` | Bearer | 既読にする |
| POST | `/api/v1/notifications/read-all` | Bearer | 全て既読 |

### 3.2 レスポンス形式

```json
{ "data": { ... } }
{ "data": [...], "meta": { "next_cursor": "...", "has_next": true } }
{ "error": { "code": "not_found", "message": "..." } }
```

## 4. データベース設計

| テーブル | 主要カラム |
|---------|-----------|
| `users` | id, provider, provider_id, email, display_name, avatar_url |
| `projects` | id, name, description, owner_id |
| `issues` | id, project_id, title, body, status(open/in_progress/completed/closed), ai_session_id, ai_result |
| `notifications` | id, user_id, issue_id, type, title, message, read |
| `ai_jobs` | id, issue_id, status(pending/running/completed/failed), attempts, max_attempts, error_msg |

## 5. 技術スタック

| 領域 | 技術 | 選定理由 |
|------|------|---------|
| Backend | Go + chi | 高パフォーマンス、stdlib 互換ルーター |
| DB Driver | pgx + sqlx | Pure Go PostgreSQL ドライバ、構造体スキャン |
| Migration | golang-migrate | CLI + Go API 対応 |
| Auth | golang-jwt + x/oauth2 | JWT 標準ライブラリ + 公式 OAuth2 |
| Web | React + Vite + Tailwind | 高速開発、TanStack Query でサーバー状態管理 |
| Mobile | React Native + Expo | クロスプラットフォーム、Web とコード共有 |
| DB | PostgreSQL 16 | 信頼性、ENUM・部分インデックス・FOR UPDATE SKIP LOCKED |
| CI/CD | GitHub Actions | PR レビューコメントに Claude Code で自動対応 |

## 6. 開発マイルストーン

- **Phase 1**: プロジェクト基盤（Go module, ミドルウェア, マイグレーション, Taskfile） ✅
- **Phase 2**: 認証システム（Google/GitHub OAuth, JWT） ✅
- **Phase 3**: Project & Issue CRUD
- **Phase 4**: AI 連携（Claude Code サブプロセス実行、ジョブキュー）
- **Phase 5**: 通知システム（In-App + Webhook）
- **Phase 6**: Web フロントエンド（React）
- **Phase 7**: モバイル（React Native + Expo）

## 7. セットアップ

```bash
# 依存関係
brew install go task golang-migrate

# データベース起動
task db:up

# マイグレーション実行
task migrate:up

# サーバー起動
task run

# テスト
task test
```
