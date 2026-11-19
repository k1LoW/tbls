package mssql

import "testing"

func TestConvertColumnType(t *testing.T) {
	tests := []struct {
		name      string
		typ       string
		maxLength int
		precision int
		scale     int
		want      string
	}{
		{"varchar", "varchar", 255, 0, 0, "varchar(255)"},
		{"varchar(MAX)", "varchar", -1, 0, 0, "varchar(MAX)"},
		{"char", "char", 5, 0, 0, "char(5)"},
		{"binary", "binary", 16, 0, 0, "binary(16)"},
		{"nchar is counted in characters, not bytes", "nchar", 10, 0, 0, "nchar(5)"},
		{"nvarchar is counted in characters, not bytes", "nvarchar", 40, 0, 0, "nvarchar(20)"},
		{"nvarchar(MAX)", "nvarchar", -1, 0, 0, "nvarchar(MAX)"},
		{"varbinary", "varbinary", 64, 0, 0, "varbinary(64)"},
		{"varbinary(MAX)", "varbinary", -1, 0, 0, "varbinary(MAX)"},
		{"decimal", "decimal", 5, 9, 2, "decimal(9,2)"},
		{"numeric", "numeric", 9, 18, 4, "numeric(18,4)"},
		{"datetime2", "datetime2", 7, 23, 3, "datetime2(3)"},
		{"datetime2 with the default scale", "datetime2", 8, 27, 7, "datetime2(7)"},
		{"time", "time", 3, 8, 0, "time(0)"},
		{"datetimeoffset", "datetimeoffset", 8, 29, 2, "datetimeoffset(2)"},
		{"datetime has a fixed scale and takes no argument", "datetime", 8, 23, 3, "datetime"},
		{"int", "int", 4, 10, 0, "int"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertColumnType(tt.typ, tt.maxLength, tt.precision, tt.scale); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
