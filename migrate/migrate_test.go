package migrate

import (
	"testing"

	"github.com/gooolib/migration/config"
	"github.com/stretchr/testify/assert"
)

func TestMigration_Versions(t *testing.T) {
	type fields struct {
		CurrentVersion string
		UpFiles        []MigrationFile
		DownFiles      []MigrationFile
		repo           *repository
		config         *config.Config
	}
	tests := []struct {
		name   string
		fields fields
		want   []string
	}{
		{
			name: "no files",
			fields: fields{
				UpFiles: []MigrationFile{},
			},
			want: []string{},
		},
		{
			name: "multiple files",
			fields: fields{
				UpFiles: []MigrationFile{
					{Path: "20230101_init.up.sql"},
					{Path: "20230201_add_users.up.sql"},
					{Path: "20230301_add_orders.up.sql"},
				},
			},
			want: []string{"20230101", "20230201", "20230301"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Migration{
				CurrentVersion: tt.fields.CurrentVersion,
				UpFiles:        tt.fields.UpFiles,
				DownFiles:      tt.fields.DownFiles,
				repo:           tt.fields.repo,
				config:         tt.fields.config,
			}
			got := m.Versions()

			assert.Equal(t, tt.want, got)
		})
	}
}

type mockStatusGetter struct {
	version string
	err     error
}

func (m *mockStatusGetter) GetCurrentVersion() (string, error) {
	return m.version, m.err
}

func TestMigration_Load(t *testing.T) {
	m := &Migration{
		statusGetter: &mockStatusGetter{},
		config:       &config.Config{},
	}

	// INFO: put ../ because current dir in test is in migrate package
	if err := m.Load("../db/migrations"); err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	assert.Equal(t, 4, len(m.UpFiles), "expected 4 up files")
	assert.Equal(t, 4, len(m.DownFiles), "expected 4 down files")
	assert.Equal(t, "20250830133803", m.UpFiles[0].Version())
	assert.Equal(t, "20250830133803", m.DownFiles[0].Version())
	assert.Equal(t, "20250830133956", m.UpFiles[3].Version())
	assert.Equal(t, "20250830133956", m.DownFiles[3].Version())
}
