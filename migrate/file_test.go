package migrate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMigrationFile_IsUp(t *testing.T) {
	tests := []struct {
		name string
		file MigrationFile
		want bool
	}{
		{
			name: "up file",
			file: MigrationFile{
				Path: "20230101_init.up.sql",
				Kind: "up",
			},
			want: true,
		},
		{
			name: "down file",
			file: MigrationFile{
				Path: "20230101_init.down.sql",
				Kind: "down",
			},
			want: false,
		},
		{
			name: "empty kind",
			file: MigrationFile{
				Path: "20230101_init.sql",
				Kind: "",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.file.IsUp()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMigrationFile_IsDown(t *testing.T) {
	tests := []struct {
		name string
		file MigrationFile
		want bool
	}{
		{
			name: "down file",
			file: MigrationFile{
				Path: "20230101_init.down.sql",
				Kind: "down",
			},
			want: true,
		},
		{
			name: "up file",
			file: MigrationFile{
				Path: "20230101_init.up.sql",
				Kind: "up",
			},
			want: false,
		},
		{
			name: "empty kind",
			file: MigrationFile{
				Path: "20230101_init.sql",
				Kind: "",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.file.IsDown()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMigrationFile_Version(t *testing.T) {
	tests := []struct {
		name string
		file MigrationFile
		want string
	}{
		{
			name: "standard migration file",
			file: MigrationFile{
				Path: "20230101_init.up.sql",
			},
			want: "20230101",
		},
		{
			name: "migration file with complex name",
			file: MigrationFile{
				Path: "20230201_add_users_table.up.sql",
			},
			want: "20230201",
		},
		{
			name: "file with full path",
			file: MigrationFile{
				Path: "/path/to/migrations/20230301_add_orders.down.sql",
			},
			want: "20230301",
		},
		{
			name: "file with no underscore",
			file: MigrationFile{
				Path: "20230401.up.sql",
			},
			want: "",
		},
		{
			name: "file with only one part",
			file: MigrationFile{
				Path: "migration.sql",
			},
			want: "",
		},
		{
			name: "empty path",
			file: MigrationFile{
				Path: "",
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.file.Version()
			assert.Equal(t, tt.want, got)
		})
	}
}