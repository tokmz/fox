package datascope

import "testing"

func TestIsValidColumn(t *testing.T) {
	tests := []struct {
		name   string
		column string
		want   bool
	}{
		{"valid dept_id", "dept_id", true},
		{"valid created_by", "created_by", true},
		{"valid user_id", "user_id", true},
		{"invalid column", "invalid_column", false},
		{"sql injection attempt", "dept_id; DROP TABLE users--", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidColumn(tt.column); got != tt.want {
				t.Errorf("IsValidColumn(%q) = %v, want %v", tt.column, got, tt.want)
			}
		})
	}
}

func TestIsValidIdentifier(t *testing.T) {
	tests := []struct {
		name       string
		identifier string
		want       bool
	}{
		{"valid table name", "users", true},
		{"valid alias", "u", true},
		{"valid with underscore", "sys_user", true},
		{"valid with number", "table123", true},
		{"empty string", "", true}, // 空字符串表示不使用别名
		{"starts with number", "123table", false},
		{"contains space", "user table", false},
		{"contains special char", "user-table", false},
		{"sql injection", "users; DROP TABLE--", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidIdentifier(tt.identifier); got != tt.want {
				t.Errorf("IsValidIdentifier(%q) = %v, want %v", tt.identifier, got, tt.want)
			}
		})
	}
}
