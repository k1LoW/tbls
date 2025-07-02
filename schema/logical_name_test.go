package schema

import (
	"database/sql"
	"strings"
	"testing"
)

func TestColumn_SetLogicalNameFromComment(t *testing.T) {
	tests := []struct {
		name              string
		column            Column
		delimiter         string
		fallbackToName    bool
		wantLogicalName   string
		wantComment       string
	}{
		{
			name: "コメントから論理名とコメントを分離",
			column: Column{
				Name:    "user_id",
				Comment: "ユーザーID|ユーザーの一意識別子",
			},
			delimiter:       "|",
			fallbackToName:  false,
			wantLogicalName: "ユーザーID",
			wantComment:     "ユーザーの一意識別子",
		},
		{
			name: "区切り文字なし、フォールバック無効",
			column: Column{
				Name:    "user_id",
				Comment: "ユーザーID",
			},
			delimiter:       "|",
			fallbackToName:  false,
			wantLogicalName: "ユーザーID",
			wantComment:     "",
		},
		{
			name: "空のコメント、フォールバック有効",
			column: Column{
				Name:    "user_id",
				Comment: "",
			},
			delimiter:       "|",
			fallbackToName:  true,
			wantLogicalName: "user_id",
			wantComment:     "",
		},
		{
			name: "空のコメント、フォールバック無効",
			column: Column{
				Name:    "user_id",
				Comment: "",
			},
			delimiter:       "|",
			fallbackToName:  false,
			wantLogicalName: "",
			wantComment:     "",
		},
		{
			name: "複数の区切り文字（最初のみで分割）",
			column: Column{
				Name:    "user_id",
				Comment: "ユーザーID|主キー|自動採番",
			},
			delimiter:       "|",
			fallbackToName:  false,
			wantLogicalName: "ユーザーID",
			wantComment:     "主キー|自動採番",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			column := tt.column // コピーして元を変更しないようにする
			column.SetLogicalNameFromComment(tt.delimiter, tt.fallbackToName)

			if column.LogicalName != tt.wantLogicalName {
				t.Errorf("SetLogicalNameFromComment() LogicalName = %v, want %v", column.LogicalName, tt.wantLogicalName)
			}
			if column.Comment != tt.wantComment {
				t.Errorf("SetLogicalNameFromComment() Comment = %v, want %v", column.Comment, tt.wantComment)
			}
		})
	}
}

func TestColumn_GetLogicalNameOrFallback(t *testing.T) {
	tests := []struct {
		name           string
		column         Column
		fallbackToName bool
		want           string
	}{
		{
			name: "論理名が設定されている場合",
			column: Column{
				Name:        "user_id",
				LogicalName: "ユーザーID",
			},
			fallbackToName: true,
			want:           "ユーザーID",
		},
		{
			name: "論理名が空、フォールバック有効",
			column: Column{
				Name:        "user_id",
				LogicalName: "",
			},
			fallbackToName: true,
			want:           "user_id",
		},
		{
			name: "論理名が空、フォールバック無効",
			column: Column{
				Name:        "user_id",
				LogicalName: "",
			},
			fallbackToName: false,
			want:           "",
		},
		{
			name: "論理名が設定されている場合（フォールバック無効でも論理名を返す）",
			column: Column{
				Name:        "user_id",
				LogicalName: "ユーザーID",
			},
			fallbackToName: false,
			want:           "ユーザーID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.column.GetLogicalNameOrFallback(tt.fallbackToName)
			if got != tt.want {
				t.Errorf("GetLogicalNameOrFallback() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestColumn_HasLogicalName(t *testing.T) {
	tests := []struct {
		name   string
		column Column
		want   bool
	}{
		{
			name: "論理名が設定されている場合",
			column: Column{
				Name:        "user_id",
				LogicalName: "ユーザーID",
			},
			want: true,
		},
		{
			name: "論理名が空の場合",
			column: Column{
				Name:        "user_id",
				LogicalName: "",
			},
			want: false,
		},
		{
			name: "論理名が空白のみの場合",
			column: Column{
				Name:        "user_id",
				LogicalName: "   ",
			},
			want: true, // 空白も文字として扱う
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.column.HasLogicalName()
			if got != tt.want {
				t.Errorf("HasLogicalName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestColumn_LogicalNameSerialization(t *testing.T) {
	// JSONシリアライゼーションのテスト
	column := Column{
		Name:        "user_id",
		Type:        "integer",
		Nullable:    false,
		Default:     sql.NullString{String: "", Valid: false},
		Comment:     "ユーザーの識別子",
		LogicalName: "ユーザーID",
	}

	// MarshalJSONをテスト
	jsonData, err := column.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON() error = %v", err)
	}

	// JSON文字列に論理名が含まれているかチェック
	jsonStr := string(jsonData)
	if !strings.Contains(jsonStr, "logical_name") {
		t.Errorf("JSON should contain logical_name field, got: %s", jsonStr)
	}
	if !strings.Contains(jsonStr, "ユーザーID") {
		t.Errorf("JSON should contain logical name value, got: %s", jsonStr)
	}
}
