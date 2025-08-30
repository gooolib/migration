package migrate

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gooolib/migration/config"
)

type Migration struct {
	CurrentVersion string
	UpFiles        []MigrationFile
	DownFiles      []MigrationFile
	repo           *repository
	config         *config.Config
}

func (m *Migration) Config() *config.Config {
	return m.config
}

func (m *Migration) Versions() []string {
	versions := make([]string, 0, len(m.UpFiles))
	for _, file := range m.UpFiles {
		versions = append(versions, file.Version())
	}
	return versions
}

func (m *Migration) Up() error {
	tx, err := m.repo.DB().Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	currentVersion, err := m.repo.GetCurrentVersion()
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	for _, file := range m.UpFiles {
		applied := file.Version() <= currentVersion
		if applied {
			continue
		}
		if err := m.executeFile(tx, file); err != nil {
			return err
		}

		if err := m.repo.RecordMigration(tx, file.Version()); err != nil {
			return fmt.Errorf("failed to record migration: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (m *Migration) Down() error {
	tx, err := m.repo.DB().Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	file := m.DownFiles[len(m.DownFiles)-1]
	if err := m.executeFile(tx, file); err != nil {
		return err
	}
	if err := m.repo.RemoveMigrationRecord(tx, file.Version()); err != nil {
		return fmt.Errorf("failed to remove migration record: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (m *Migration) DownAll() error {
	tx, err := m.repo.DB().Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	currentVersion, err := m.repo.GetCurrentVersion()
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	for i := len(m.DownFiles) - 1; i >= 0; i-- {
		file := m.DownFiles[i]
		applied := file.Version() > currentVersion
		if applied {
			continue
		}

		if err := m.executeFile(tx, file); err != nil {
			return err
		}

		if err := m.repo.RemoveMigrationRecord(tx, file.Version()); err != nil {
			return fmt.Errorf("failed to remove migration record: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (m *Migration) SoftReset() error {
	if err := m.DownAll(); err != nil {
		return fmt.Errorf("failed to reset migrations: %w", err)
	}

	if err := m.Up(); err != nil {
		return fmt.Errorf("failed to reapply migrations: %w", err)
	}

	return nil
}

func (m *Migration) HardReset() error {
	return m.repo.ResetMigrations()
}

func (m *Migration) executeFile(tx *sql.Tx, file MigrationFile) error {
	content, err := ioutil.ReadFile(file.Path)
	if err != nil {
		return fmt.Errorf("failed to read migration file %s: %w", file.Path, err)
	}

	if tx == nil {
		if m.repo.DB().Exec(string(content)); err != nil {
			return fmt.Errorf("failed to execute migration file %s: %w", file.Path, err)
		}
		log.Printf("Successfully executed migration: %s", file.Path)
	} else {
		if _, err := tx.Exec(string(content)); err != nil {
			return fmt.Errorf("failed to execute migration file %s: %w", file.Path, err)
		}
	}

	log.Printf("Successfully executed migration in transaction: %s", file.Path)
	return nil
}

func (m *Migration) Load(path string) error {
	files, err := filepath.Glob(filepath.Join(path, "*.sql"))
	if err != nil {
		return fmt.Errorf("failed to list migration files: %w", err)
	}

	sort.Strings(files)

	for _, file := range files {
		fileName := filepath.Base(file)
		if strings.Contains(fileName, "up") {
			m.UpFiles = append(m.UpFiles, MigrationFile{
				Path: file,
				Kind: "up",
			})
		} else if strings.Contains(fileName, "down") {
			m.DownFiles = append(m.DownFiles, MigrationFile{
				Path: file,
				Kind: "down",
			})
		} else {
			return fmt.Errorf("invalid migration file name: %s, expected 'up' or 'down' in the name", fileName)
		}
	}

	version, err := m.repo.GetCurrentVersion()
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	m.CurrentVersion = version

	return nil
}

func (m *Migration) RunSingleUp(file MigrationFile) error {
	if err := m.executeFile(nil, file); err != nil {
		return err
	}

	if err := m.repo.RecordMigration(nil, file.Version()); err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	return nil
}

func (m *Migration) RunSingleDown(file MigrationFile) error {
	if err := m.executeFile(nil, file); err != nil {
		return err
	}

	if err := m.repo.RemoveMigrationRecord(nil, file.Version()); err != nil {
		return fmt.Errorf("failed to remove migration record: %w", err)
	}

	return nil
}

func (m *Migration) GetCurrentVersion() string {
	version, err := m.repo.GetCurrentVersion()
	if err != nil {
		log.Printf("Error getting current version: %v", err)
		return ""
	}
	return version
}

func (m *Migration) CreateMigrationTable() error {
	return m.repo.CreateMigrationTable()
}

func NewMigration(config *config.Config) (*Migration, error) {
	repo, err := newRepository(config.Database.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to create repository: %w", err)
	}

	migration := &Migration{
		repo:   repo,
		config: config,
	}

	if err := migration.CreateMigrationTable(); err != nil {
		return nil, err
	}

	return migration, nil
}

func (m *Migration) Status() ([]SchemaMigrationStatus, error) {
	applied, err := m.repo.ListAppliedMigrations()
	if err != nil {
		return nil, fmt.Errorf("failed to list applied migrations: %w", err)
	}

	statuses := make([]SchemaMigrationStatus, len(m.Versions()))
	for i, version := range m.Versions() {
		var found *SchemaMigration
		for _, v := range applied {
			if v.Version == version {
				found = &v
				break
			}
		}

		if found != nil {
			statuses[i] = SchemaMigrationStatus{
				Version:   version,
				AppliedAt: &found.AppliedAt,
				Status:    "up",
			}
		} else {
			statuses[i] = SchemaMigrationStatus{
				Version:   version,
				AppliedAt: nil,
				Status:    "pending",
			}
		}
	}

	return statuses, nil
}

func (m *Migration) Close() error {
	if m.repo != nil {
		return m.repo.Close()
	}
	return nil
}

func (m *Migration) FindFileByVersion(version string, migrationType string) *MigrationFile {
	if migrationType == "up" {
		return findInFiles(m.UpFiles, version)
	}

	if migrationType == "down" {
		return findInFiles(m.DownFiles, version)
	}
	return nil
}

func (m *Migration) IsMigrationApplied(version string) (bool, error) {
	return m.repo.IsMigrationApplied(version)
}

func findInFiles(files []MigrationFile, version string) *MigrationFile {
	for _, file := range files {
		if file.Version() == version {
			return &file
		}
	}
	return nil
}
