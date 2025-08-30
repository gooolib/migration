package command

import (
	"flag"

	"github.com/gooolib/migration/internal/migrate"
)

type RollbackCommand struct {
	args      *flag.FlagSet
	migration *migrate.Migration
}

func (c *RollbackCommand) Exec() error {
	return c.migration.Down()
}

func (c *RollbackCommand) ParseArgs() error {
	return nil
}
