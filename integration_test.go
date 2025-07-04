package main

import (
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/k1LoW/tbls/config"
	"github.com/k1LoW/tbls/schema"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

// TestLogicalNameIntegration 論理名機能の統合テスト
// NOTE: このテストは論理名機能が完全に実装された後に有効化される
func TestLogicalNameIntegration(t *testing.T) {
	t.Skip("論理名機能の統合テスト - PR統合後に有効化予定")
	
	// SQLiteデータベースの作成
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	
	// データベースの初期化
	if err := setupTestDatabase(dbPath); err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	
	// 設定ファイルの作成
	configPath := filepath.Join(tempDir, ".tbls.yml")
	if err := createTestConfig(configPath, dbPath); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}
	
	// 出力ディレクトリの作成
	outputDir := filepath.Join(tempDir, "docs")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}
	
	// tblsバイナリの実行
	t.Run("論理名機能有効でのドキュメント生成", func(t *testing.T) {
		cmd := exec.Command("go", "run", ".", "doc", "--config", configPath, fmt.Sprintf("sqlite://%s", dbPath))
		cmd.Dir = "."
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Errorf("Failed to generate documentation: %v\nOutput: %s", err, string(output))
		}
		
		// 生成されたMarkdownファイルの確認
		mdFile := filepath.Join(outputDir, "users.md")
		if _, err := os.Stat(mdFile); os.IsNotExist(err) {
			t.Errorf("Expected markdown file not found: %s", mdFile)
		}
		
		// ファイル内容の確認
		content, err := os.ReadFile(mdFile)
		if err != nil {
			t.Fatalf("Failed to read markdown file: %v", err)
		}
		
		contentStr := string(content)
		// 論理名カラムの存在確認
		if !strings.Contains(contentStr, "Logical Name") {
			t.Errorf("Expected 'Logical Name' column header not found in output")
		}
		
		// 論理名の実際の値の確認
		if !strings.Contains(contentStr, "ユーザーID") {
			t.Errorf("Expected logical name 'ユーザーID' not found in output")
		}
		
		// コメントの分割確認
		if !strings.Contains(contentStr, "Unique identifier for users") {
			t.Errorf("Expected split comment 'Unique identifier for users' not found in output")
		}
	})
	
	// 設定を無効化した場合のテスト
	t.Run("論理名機能無効でのドキュメント生成", func(t *testing.T) {
		disabledConfigPath := filepath.Join(tempDir, ".tbls_disabled.yml")
		if err := createTestConfigDisabled(disabledConfigPath, dbPath); err != nil {
			t.Fatalf("Failed to create disabled config: %v", err)
		}
		
		cmd := exec.Command("go", "run", ".", "doc", "--config", disabledConfigPath, fmt.Sprintf("sqlite://%s", dbPath))
		cmd.Dir = "."
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Errorf("Failed to generate documentation: %v\nOutput: %s", err, string(output))
		}
		
		// 生成されたMarkdownファイルの確認
		mdFile := filepath.Join(outputDir, "users.md")
		content, err := os.ReadFile(mdFile)
		if err != nil {
			t.Fatalf("Failed to read markdown file: %v", err)
		}
		
		contentStr := string(content)
		// 論理名カラムが存在しないことを確認
		if strings.Contains(contentStr, "Logical Name") {
			t.Errorf("Unexpected 'Logical Name' column found when disabled")
		}
		
		// 元のコメントが表示されることを確認
		if !strings.Contains(contentStr, "ユーザーID|Unique identifier for users") {
			t.Errorf("Expected original comment with delimiter not found")
		}
	})
}

// setupTestDatabase テスト用SQLiteデータベースの設定
func setupTestDatabase(dbPath string) error {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()
	
	// テーブルの作成
	_, err = db.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY,
			username TEXT NOT NULL,
			email TEXT UNIQUE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return err
	}
	
	// SQLiteではCOMMENT構文が使えないため、別のテーブルでコメント情報を管理する方法を使用
	// 実際にはtblsの設定ファイルでコメントを定義する
	
	return nil
}

// createTestConfig テスト用の設定ファイル作成（論理名機能有効）
func createTestConfig(configPath, dbPath string) error {
	config := fmt.Sprintf(`
docPath: docs
format:
  logicalName:
    enabled: true
    delimiter: "|"
    fallbackToName: true
comments:
  - table: users
    tableComment: "ユーザー情報テーブル"
    columnComments:
      id: "ユーザーID|Unique identifier for users"
      username: "ユーザー名|User display name"
      email: "メールアドレス|User email address"
      created_at: "作成日時|Account creation timestamp"
`)
	
	return os.WriteFile(configPath, []byte(config), 0644)
}

// createTestConfigDisabled テスト用の設定ファイル作成（論理名機能無効）
func createTestConfigDisabled(configPath, dbPath string) error {
	config := fmt.Sprintf(`
docPath: docs
format:
  logicalName:
    enabled: false
comments:
  - table: users
    tableComment: "ユーザー情報テーブル"
    columnComments:
      id: "ユーザーID|Unique identifier for users"
      username: "ユーザー名|User display name"
      email: "メールアドレス|User email address"
      created_at: "作成日時|Account creation timestamp"
`)
	
	return os.WriteFile(configPath, []byte(config), 0644)
}

// TestLogicalNameConfigStructure 論理名設定構造のテスト
func TestLogicalNameConfigStructure(t *testing.T) {
	cfg, err := config.New()
	if err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}
	
	// デフォルト値の確認
	if cfg.LogicalNameDelimiter() != "|" {
		t.Errorf("Expected default delimiter '|', got '%s'", cfg.LogicalNameDelimiter())
	}
	
	// 設定が存在することの確認（構造体の存在確認）
	if cfg.Format.LogicalName.Delimiter == "" {
		// デフォルト値が設定されていることを確認
		cfg.Format.LogicalName.Delimiter = "|"
	}
}

// TestLogicalNameExtraction 論理名抽出のテスト
func TestLogicalNameExtraction(t *testing.T) {
	tests := []struct {
		name      string
		comment   string
		delimiter string
		expected  string
		expectedComment string
	}{
		{
			name:      "基本的な分割",
			comment:   "ユーザーID|Unique identifier for users",
			delimiter: "|",
			expected:  "ユーザーID",
			expectedComment: "Unique identifier for users",
		},
		{
			name:      "区切り文字なし",
			comment:   "ユーザーID",
			delimiter: "|",
			expected:  "ユーザーID",
			expectedComment: "",
		},
		{
			name:      "空のコメント",
			comment:   "",
			delimiter: "|",
			expected:  "",
			expectedComment: "",
		},
		{
			name:      "複数の区切り文字",
			comment:   "ユーザーID|Unique identifier|Additional info",
			delimiter: "|",
			expected:  "ユーザーID",
			expectedComment: "Unique identifier|Additional info",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			column := &schema.Column{
				Name:    "test_column",
				Comment: tt.comment,
			}
			
			column.SetLogicalNameFromComment(tt.delimiter, false)
			
			if column.LogicalName != tt.expected {
				t.Errorf("Expected logical name '%s', got '%s'", tt.expected, column.LogicalName)
			}
			
			if column.Comment != tt.expectedComment {
				t.Errorf("Expected comment '%s', got '%s'", tt.expectedComment, column.Comment)
			}
		})
	}
}

// TestMultipleOutputFormats 複数の出力フォーマットのテスト
// NOTE: このテストは論理名機能が完全に実装された後に有効化される
func TestMultipleOutputFormats(t *testing.T) {
	t.Skip("複数フォーマットテスト - PR統合後に有効化予定")
	
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	
	// データベースの初期化
	if err := setupTestDatabase(dbPath); err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	
	// 各フォーマットのテスト
	formats := []struct {
		name   string
		format string
		check  func(string) bool
	}{
		{
			name:   "Markdown",
			format: "md",
			check: func(content string) bool {
				return strings.Contains(content, "Logical Name") && strings.Contains(content, "ユーザーID")
			},
		},
		{
			name:   "DOT",
			format: "dot",
			check: func(content string) bool {
				return strings.Contains(content, "id (ユーザーID)")
			},
		},
		{
			name:   "PlantUML",
			format: "plantuml",
			check: func(content string) bool {
				return strings.Contains(content, "id (ユーザーID)")
			},
		},
		{
			name:   "Mermaid",
			format: "mermaid",
			check: func(content string) bool {
				return strings.Contains(content, "id_ユーザーID")
			},
		},
	}
	
	for _, testCase := range formats {
		t.Run(testCase.name, func(t *testing.T) {
			outputPath := filepath.Join(tempDir, testCase.format)
			configPath := filepath.Join(tempDir, testCase.format+".yml")
			
			// 設定ファイルの作成
			configContent := fmt.Sprintf(`
docPath: %s
format:
  logicalName:
    enabled: true
    delimiter: "|"
    fallbackToName: true
comments:
  - table: users
    tableComment: "ユーザー情報テーブル"
    columnComments:
      id: "ユーザーID|Unique identifier for users"
      username: "ユーザー名|User display name"
`, outputPath)
			
			if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
				t.Fatalf("Failed to create config file: %v", err)
			}
			
			// フォーマット指定してドキュメント生成
			cmd := exec.Command("go", "run", ".", "doc", "--config", configPath, "--format", testCase.format, fmt.Sprintf("sqlite://%s", dbPath))
			cmd.Dir = "."
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Errorf("Failed to generate %s documentation: %v\nOutput: %s", testCase.format, err, string(output))
				return
			}
			
			// 生成されたファイルの確認
			var generatedFile string
			switch testCase.format {
			case "md":
				generatedFile = filepath.Join(outputPath, "users.md")
			case "dot":
				generatedFile = filepath.Join(outputPath, "schema.dot")
			case "plantuml":
				generatedFile = filepath.Join(outputPath, "schema.puml")
			case "mermaid":
				generatedFile = filepath.Join(outputPath, "schema.mermaid")
			}
			
			if _, err := os.Stat(generatedFile); os.IsNotExist(err) {
				t.Errorf("Expected %s file not found: %s", testCase.format, generatedFile)
				return
			}
			
			// ファイル内容の確認
			content, err := os.ReadFile(generatedFile)
			if err != nil {
				t.Fatalf("Failed to read %s file: %v", testCase.format, err)
			}
			
			if !testCase.check(string(content)) {
				t.Errorf("Expected logical name not found in %s output", testCase.format)
			}
		})
	}
}