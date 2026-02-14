# 設計ドキュメント

## 要件

- GitHub のような issue 管理機能
- 起票された issue は AI（Claude Code）が確認してタスクとして実行する
- issue の一覧を取得できる API の提供
- プロジェクトを作成でき、プロジェクトごとに issue を作成できる
- issue に対して完了報告ができる
- Web, モバイルで使いたい

## 決定事項

| 項目 | 決定 | 理由 |
|------|------|------|
| Backend | Go | 高パフォーマンス、既存ルール整備済み |
| Frontend Web | React + TypeScript | エコシステムの豊富さ、React Native とコード共有 |
| Frontend Mobile | React Native + Expo | クロスプラットフォーム |
| Database | PostgreSQL | 信頼性、ジョブキュー（SKIP LOCKED）対応 |
| Auth | Google/GitHub OAuth | 一人プロジェクト想定、権限管理不要 |
| AI 連携 | Claude Code CLI subprocess | Issue open → Claude Code 起動 → 完了通知 |
| Task Runner | Taskfile | Makefile の代替、YAML ベース |
| Container | compose.yaml | docker-compose.yaml は非推奨 |
| CI/CD | GitHub Actions | PR レビューコメントに Claude Code で自動対応 |

## 未決事項

- [ ] GitHub Actions の具体的なワークフロー定義
- [ ] Push 通知の実装方式（Expo Push Notifications vs Firebase）
- [ ] Claude Code 実行時のプロジェクトスコープ（作業ディレクトリの分離方法）
- [ ] 本番環境のデプロイ先（Cloud Run, Fly.io, etc.）
