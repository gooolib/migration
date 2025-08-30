package migrate

import "time"

type SchemaMigration struct {
	Version   string
	AppliedAt time.Time
}

type SchemaMigrationStatus struct {
	Version   string
	AppliedAt *time.Time
	Status    string // "up" or "pending"
}
