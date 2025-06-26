# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

このプロジェクトの担当者は、日本語でのコミュニケーションを希望しています。ドキュメント・コメントは全て日本語で記載してください。

## プロジェクト概要

tblsは、CI-Friendlyなデータベースドキュメント生成ツールです。単一バイナリで動作し、多数のデータベースをサポートしています。

## 開発コマンド

### ビルドとテスト
```bash
# バイナリのビルド
make build

# 全テストの実行（各種データベースタグ付き）
go test ./... -tags 'bq clickhouse dynamo mariadb mongodb mssql mysql postgres redshift snowflake spanner sqlite' -coverprofile=coverage.out -covermode=count

# 特定のパッケージのテスト実行
go test ./datasource -v
go test ./drivers/mysql -v

# 特定のテストケースの実行
go test -run TestAnalyzeSchema ./datasource -v

# Lintの実行
make lint
# または
golangci-lint run ./...

# CI環境での完全なテスト実行
make ci
```

### 開発用データベースのセットアップ
```bash
# テスト用データベースの起動（Docker Compose使用）
docker compose up -d

# テスト用データベースの初期化
make db

# SQLiteのみの初期化
make db_sqlite
```

### ドキュメント生成とテスト
```bash
# サンプルドキュメントの生成
make doc

# ドキュメントとデータベースの差分確認
make testdoc
```

## アーキテクチャとコード構造

### コマンド構造（cmd/）
- Cobraライブラリを使用したCLI実装
- 各サブコマンドは独立したファイルで実装
- 外部サブコマンドのサポート（PATH上の`tbls-*`実行可能ファイル）

### データベースドライバー（drivers/）
- 各データベース固有の実装を含む
- 共通インターフェース: `Driver`を実装
- 新しいドライバー追加時は`drivers.go`への登録が必要

### 出力フォーマット（output/）
- 各出力形式は独立したパッケージ
- テンプレートエンジンを使用（Go template）
- カスタムテンプレートのサポート

### スキーマ表現（schema/）
- `Schema`構造体がデータベース全体を表現
- `Table`、`Column`、`Relation`等の基本構造体
- JSONとYAMLでのシリアライズ/デシリアライズ対応

### 設定管理（config/）
- `.tbls.yml`ファイルでの設定
- 環境変数の展開サポート（`${}`構文）
- DSN、フィルタ、Lint規則等の設定

## 重要な実装パターン

### エラーハンドリング
- 基本的に早期リターンでエラーを伝播
- CLIレベルでのエラー表示とexit処理

### テストデータ
- `testdata/`ディレクトリにテスト用SQL、設定ファイル、期待値を配置
- Goldenファイルパターンの使用（`.golden`拡張子）

### データベース接続
- DSN形式: `driver://user:pass@host:port/dbname?option=value`
- 環境変数での設定も可能（`TBLS_DSN`）

### 外部コマンド/ドライバー
- PATH上の`tbls-*`実行可能ファイルを外部サブコマンドとして認識
- 標準入出力を通じた通信

## デバッグとトラブルシューティング

### 詳細ログの表示
```bash
# --debugオプションでデバッグ情報を表示
tbls doc --debug postgres://...
```

### 特定のテーブルのみ処理
```bash
# --table/-tオプションでテーブルを指定
tbls doc postgres://... -t users -t posts
```

### 設定ファイルの検証
```bash
# --config/-cで設定ファイルを明示的に指定
tbls doc --config custom.yml postgres://...
```

## 注意事項

- 新機能追加時は必ずテストを追加
- データベース固有の処理はドライバー層に実装
- 出力フォーマット固有の処理は出力層に実装
- 環境変数の展開は`${VAR_NAME}`形式を使用
- Lint機能追加時は`config/lint.go`を更新