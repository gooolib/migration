package main

	"flag"
	"fmt"
	"log"

	"github.com/gooolib/errors"
	"github.com/gooolib/migration/internal/migrate"
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

	m, err := migrate.NewMigration("postgres://postgres:postgres@127.0.0.1:5432/asset_trader_development?sslmode=disable")
	if err != nil {
		log.Fatalf("Failed to create migration: %v", err)
	}

	if err := m.Load("./db/migrations"); err != nil {
		log.Fatalf("Failed to load migrations: %v", err)
	}

	version := m.GetCurrentVersion()
	if version == "" {
		version = "initial"
	}
	fmt.Println("Current version:", version)

	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		log.Fatalf("Usage: migrate <command> args...\nAvailable commands: up, down, reset")
		return
	}

	cmd := args[0]
	if cmd == "up" {
		if err := m.Up(); err != nil {
			log.Fatalf("Failed to apply migrations: %v", err)
		}
		fmt.Println("Migrations applied successfully.")
		return
	}

	if cmd == "down" {
		if err := m.Down(); err != nil {
			log.Fatalf("Failed to revert migrations: %v", err)
		}
		fmt.Println("Migrations reverted successfully.")
		return
	}

	if cmd == "reset" {
		if err := m.HardReset(); err != nil {
			log.Fatalf("Failed to reset migrations: %v", err)
		}
		fmt.Println("Migrations reset successfully.")
		return
	}

	if cmd == "generate" {
		if len(args) < 2 {
			log.Fatalf("Usage: migrate generate <migration_name>")
			return
		}
		migrationName := args[1]
		m.Generate("./db/migrations", migrationName)
	}
}

