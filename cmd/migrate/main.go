package main

import (
	"log"

	"github.com/gooolib/errors"
	"github.com/gooolib/migration/command"
	"github.com/gooolib/migration/config"
	"github.com/gooolib/migration/migrate"
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			err, ok := err.(errors.Error)
			if ok {
				log.Printf("Error: %s\nStack trace:\n%s", err.Error(), err.StackTrace())
			} else {
				e := errors.Wrap(err)
				log.Printf("Error: %s\nStack trace:\n%s", e.Error(), e.StackTrace())
			}
			panic(err)
		}
	}()

	cfg := &config.Config{
		Database: config.DBConfig{
			Dialect:  "postgres",
			Host:     "127.0.0.1",
			Port:     5432,
			Username: "postgres",
			Password: "postgres",
			Database: "gooolib_migration_development",
			SSLMode:  "disable",
		},
		Command: config.NewCmdConfig(""),
	}

	m, err := migrate.NewMigration(cfg)
	if err != nil {
		log.Fatalf("Failed to create migration: %v", err)
	}

	if err := m.Load(cfg.Command.MigrationDir); err != nil {
		log.Fatalf("Failed to load migrations: %v", err)
	}

	cmd, err := command.NewCommand(m)
	if err != nil {
		log.Fatalf("Failed to parse command: %v", err)
	}

	if err := cmd.Exec(); err != nil {
		log.Fatalf("Command execution failed: %v", err)
	}
}
