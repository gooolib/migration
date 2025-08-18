package migrate

import (
	"path/filepath"
	"strings"
)

type MigrationFile struct {
	Path string
	Kind string
}

func (mf *MigrationFile) IsUp() bool {
	return mf.Kind == "up"
}

func (mf *MigrationFile) IsDown() bool {
	return mf.Kind == "down"
}

func (mf *MigrationFile) Version() string {
	parts := strings.Split(filepath.Base(mf.Path), "_")
	if len(parts) < 2 {
		return ""
	}
	return parts[0]
}
