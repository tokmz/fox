package datascope

import "regexp"

var identifierRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

// IsValidColumn 验证字段名是否合法（防止 SQL 注入）
func IsValidColumn(column string) bool {
	return ValidColumns[column]
}

// IsValidIdentifier 验证标识符是否合法（表名、别名等）
// 仅允许字母、数字、下划线，且必须以字母或下划线开头
func IsValidIdentifier(s string) bool {
	if s == "" {
		return true // 空字符串表示不使用表别名
	}
	return identifierRegex.MatchString(s)
}
