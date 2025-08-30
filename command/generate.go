package command

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/gooolib/migration/migrate"
)

type GenerateCommand struct {
	migration *migrate.Migration
	args      *flag.FlagSet
	Name      string
}

func (c *GenerateCommand) Exec() error {
	timestamp := time.Now().Format("20060102150405")
	upfileName := fmt.Sprintf("%s_%s.up.sql", timestamp, c.Name)
	downfileName := fmt.Sprintf("%s_%s.down.sql", timestamp, c.Name)

	config := c.migration.Config()
	migrationDir := config.Command.MigrationDir
	files := []string{upfileName, downfileName}
	filePaths := make([]string, len(files))
	for i, fileName := range files {
		filePath := filepath.Join(migrationDir, fileName)
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

func (c *GenerateCommand) ParseArgs() error {
	c.Name = c.args.Arg(0)
	if c.Name == "" {
		return fmt.Errorf("migration name is required")
	}

	return nil
}
