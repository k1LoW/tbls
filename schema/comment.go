package schema

import "strings"

// SplitCommentParts コメントを論理名とコメント部分に分割します
type SplitCommentParts struct {
	LogicalName string
	Comment     string
}

// SplitComment は指定された区切り文字でコメントを論理名とコメント部分に分割します
// 区切り文字が見つからない場合、全体を論理名として扱います
func SplitComment(comment string, delimiter string) SplitCommentParts {
	if comment == "" {
		return SplitCommentParts{}
	}

	parts := strings.SplitN(comment, delimiter, 2)
	result := SplitCommentParts{
		LogicalName: strings.TrimSpace(parts[0]),
	}

	if len(parts) > 1 {
		result.Comment = strings.TrimSpace(parts[1])
	}

	return result
}

// ExtractLogicalName はコメントから論理名を抽出します
// フォールバック設定により、論理名が空の場合は物理名を返します
func ExtractLogicalName(comment string, delimiter string, physicalName string, fallbackToName bool) string {
	parts := SplitComment(comment, delimiter)
	
	if parts.LogicalName != "" {
		return parts.LogicalName
	}
	
	if fallbackToName {
		return physicalName
	}
	
	return ""
}

// ExtractCleanComment はコメントから論理名部分を除去した純粋なコメント部分を返します
func ExtractCleanComment(comment string, delimiter string) string {
	parts := SplitComment(comment, delimiter)
	return parts.Comment
}
