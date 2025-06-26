package schema

import "testing"

func TestSplitComment(t *testing.T) {
	tests := []struct {
		name      string
		comment   string
		delimiter string
		want      SplitCommentParts
	}{
		{
			name:      "コメント分割（区切り文字あり）",
			comment:   "ユーザーID|ユーザーの一意識別子",
			delimiter: "|",
			want: SplitCommentParts{
				LogicalName: "ユーザーID",
				Comment:     "ユーザーの一意識別子",
			},
		},
		{
			name:      "コメント分割（区切り文字なし）",
			comment:   "ユーザーID",
			delimiter: "|",
			want: SplitCommentParts{
				LogicalName: "ユーザーID",
				Comment:     "",
			},
		},
		{
			name:      "空のコメント",
			comment:   "",
			delimiter: "|",
			want: SplitCommentParts{
				LogicalName: "",
				Comment:     "",
			},
		},
		{
			name:      "スペース含みのコメント",
			comment:   " ユーザー名 | ユーザーの表示名  ",
			delimiter: "|",
			want: SplitCommentParts{
				LogicalName: "ユーザー名",
				Comment:     "ユーザーの表示名",
			},
		},
		{
			name:      "複数の区切り文字（最初のみ分割）",
			comment:   "ユーザーID|主キー|自動採番",
			delimiter: "|",
			want: SplitCommentParts{
				LogicalName: "ユーザーID",
				Comment:     "主キー|自動採番",
			},
		},
		{
			name:      "異なる区切り文字",
			comment:   "ユーザーID;ユーザーの一意識別子",
			delimiter: ";",
			want: SplitCommentParts{
				LogicalName: "ユーザーID",
				Comment:     "ユーザーの一意識別子",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SplitComment(tt.comment, tt.delimiter)
			if got.LogicalName != tt.want.LogicalName {
				t.Errorf("SplitComment().LogicalName = %v, want %v", got.LogicalName, tt.want.LogicalName)
			}
			if got.Comment != tt.want.Comment {
				t.Errorf("SplitComment().Comment = %v, want %v", got.Comment, tt.want.Comment)
			}
		})
	}
}

func TestExtractLogicalName(t *testing.T) {
	tests := []struct {
		name           string
		comment        string
		delimiter      string
		physicalName   string
		fallbackToName bool
		want           string
	}{
		{
			name:           "論理名あり、フォールバック有効",
			comment:        "ユーザーID|識別子",
			delimiter:      "|",
			physicalName:   "user_id",
			fallbackToName: true,
			want:           "ユーザーID",
		},
		{
			name:           "論理名なし、フォールバック有効",
			comment:        "",
			delimiter:      "|",
			physicalName:   "user_id",
			fallbackToName: true,
			want:           "user_id",
		},
		{
			name:           "論理名なし、フォールバック無効",
			comment:        "",
			delimiter:      "|",
			physicalName:   "user_id",
			fallbackToName: false,
			want:           "",
		},
		{
			name:           "区切り文字なしの論理名",
			comment:        "ユーザーID",
			delimiter:      "|",
			physicalName:   "user_id",
			fallbackToName: true,
			want:           "ユーザーID",
		},
		{
			name:           "空白のみの論理名、フォールバック有効",
			comment:        " | コメント",
			delimiter:      "|",
			physicalName:   "user_id",
			fallbackToName: true,
			want:           "user_id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractLogicalName(tt.comment, tt.delimiter, tt.physicalName, tt.fallbackToName)
			if got != tt.want {
				t.Errorf("ExtractLogicalName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractCleanComment(t *testing.T) {
	tests := []struct {
		name      string
		comment   string
		delimiter string
		want      string
	}{
		{
			name:      "論理名付きコメント",
			comment:   "ユーザーID|ユーザーの一意識別子",
			delimiter: "|",
			want:      "ユーザーの一意識別子",
		},
		{
			name:      "論理名のみ",
			comment:   "ユーザーID",
			delimiter: "|",
			want:      "",
		},
		{
			name:      "空のコメント",
			comment:   "",
			delimiter: "|",
			want:      "",
		},
		{
			name:      "複数区切り文字",
			comment:   "ユーザーID|主キー|自動採番",
			delimiter: "|",
			want:      "主キー|自動採番",
		},
		{
			name:      "スペース処理",
			comment:   " ユーザー名 | ユーザーの表示名 ",
			delimiter: "|",
			want:      "ユーザーの表示名",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractCleanComment(tt.comment, tt.delimiter)
			if got != tt.want {
				t.Errorf("ExtractCleanComment() = %v, want %v", got, tt.want)
			}
		})
	}
}