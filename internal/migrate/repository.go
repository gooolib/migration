package migrate

import (
	"database/sql"
	"fmt"

	"github.com/gooolib/errors"
	_ "github.com/lib/pq"
)

type repository struct {
	db *sql.DB
}

func (r *repository) DB() *sql.DB {
	return r.db
}

func newRepository(dbURL string) (*repository, error) {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &repository{db: db}, nil
}

func (r *repository) Close() error {
	if r.db != nil {
		return r.db.Close()
	}
	return nil
}

func (r *repository) CreateMigrationTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		version VARCHAR(255) PRIMARY KEY,
		applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`

	_, err := r.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create migration table: %w", err)
	}

	return nil
}

func (r *repository) GetCurrentVersion() (string, error) {
	query := "SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 1"
	var version string
	err := r.db.QueryRow(query).Scan(&version)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", errors.Errorf("error getting current version: %w", err)
	}
	return version, nil
}

func (r *repository) RecordMigration(tx *sql.Tx, version string) error {
	query := "INSERT INTO schema_migrations (version) VALUES ($1)"
	if err := r.execQuery(tx, query, version); err != nil {
		return errors.Wrap(err)
	}
	return nil
}

func (r *repository) ExistMigrationRecord(tx *sql.Tx, version string) (bool, error) {
	query := "SELECT version FROM schema_migrations WHERE version = $1 LIMIT 1"
	result := ""
	if err := tx.QueryRow(query, version).Scan(&result); err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, errors.Wrap(err)
	}

	return true, nil
}

func (r *repository) RemoveMigrationRecord(tx *sql.Tx, version string) error {
	exist, err := r.ExistMigrationRecord(tx, version)
	if err != nil {
		return errors.Wrap(err)
	}
	if !exist {
		return nil
	}
	query := "DELETE FROM schema_migrations WHERE version = $1"
	if err := r.execQuery(tx, query, version); err != nil {
		return errors.Wrap(err)
	}
	return nil
}

func (r *repository) IsMigrationApplied(version string) (bool, error) {
	query := "SELECT COUNT(*) FROM schema_migrations WHERE version = $1"
	var count int
	err := r.db.QueryRow(query, version).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err)
	}
	return count > 0, nil
}

func (r *repository) ExecuteSQL(query string) error {
	_, err := r.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to execute SQL: %w", err)
	}
	return nil
}

func (r *repository) ResetMigrations() error {
	tx, err := r.db.Begin()
	if err != nil {
		return errors.Wrap(err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()
	// Truncate schema_migrations if exists
	_, err = tx.Exec("TRUNCATE TABLE schema_migrations")
	if err != nil {
		return fmt.Errorf("failed to truncate schema_migrations table: %w", err)
	}

	// Get all tables in the public schema
	rows, err := tx.Query(`
		SELECT tablename 
		FROM pg_tables 
		WHERE schemaname = 'public'
	`)
	if err != nil {
		return fmt.Errorf("failed to get table list: %w", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tablename string
		if err := rows.Scan(&tablename); err != nil {
			return fmt.Errorf("failed to scan table name: %w", err)
		}
		tables = append(tables, tablename)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating over tables: %w", err)
	}

	// Drop all tables
	for _, table := range tables {
		_, err := tx.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", table))
		if err != nil {
			return fmt.Errorf("failed to drop table %s: %w", table, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Recreate the migration table
	return r.CreateMigrationTable()
}

func (r *repository) ListAppliedMigrations() ([]SchemaMigration, error) {
	query := "SELECT version, applied_at FROM schema_migrations ORDER BY version"
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	defer rows.Close()

	var migrations []SchemaMigration
	for rows.Next() {
		var m SchemaMigration
		if err := rows.Scan(&m.Version, &m.AppliedAt); err != nil {
			return nil, errors.Wrap(err)
		}
		migrations = append(migrations, m)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err)
	}

	return migrations, nil
}

func (r *repository) execQuery(tx *sql.Tx, query string, args ...any) error {
	if tx == nil {
		if _, err := r.db.Exec(query, args...); err != nil {
			return errors.Wrap(err)
		}
	} else {
		if _, err := tx.Exec(query, args...); err != nil {
			return errors.Wrap(err)
		}
	}
	return nil
}
