package migrate

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type Migration struct {
	CurrentVersion string
	UpFiles        []MigrationFile
	DownFiles      []MigrationFile
	repo           *repository
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
	if err := m.Down(); err != nil {
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

func (m *Migration) Generate(path string, name string) error {
	timestamp := time.Now().Format("20060102150405")
	upfileName := fmt.Sprintf("%s_%s.up.sql", timestamp, name)
	downfileName := fmt.Sprintf("%s_%s.down.sql", timestamp, name)

	files := []string{upfileName, downfileName}
	filePaths := make([]string, len(files))
	for i, fileName := range files {
		filePath := filepath.Join(path, fileName)
		filePaths[i] = filePath

		file, err := os.Create(filePath)
		if err != nil {
			return fmt.Errorf("failed to create migration file: %w", err)
		}
		defer file.Close()
		template := fmt.Sprintf("-- Migration\n-- Created at: %s\n\n-- Write your SQL here\n", time.Now().Format("2006-01-02 15:04:05"))
		if _, err := file.WriteString(template); err != nil {
			return fmt.Errorf("failed to write migration template: %w", err)
		}
	}

	if len(filePaths) > 0 {
		log.Printf("Migration files generated successfully:")
		for _, filePath := range filePaths {
			log.Printf("file: %s", filePath)
		}
	}
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

func NewMigration(dbURL string) (*Migration, error) {
	repo, err := newRepository(dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create repository: %w", err)
	}

	migration := &Migration{
		repo: repo,
	}

	if err := migration.CreateMigrationTable(); err != nil {
		return nil, err
	}

	return migration, nil
}

func (m *Migration) Close() error {
	if m.repo != nil {
		return m.repo.Close()
	}
	return nil
}

func (m *Migration) IsMigrationApplied(version string) (bool, error) {
	return m.repo.IsMigrationApplied(version)
}
