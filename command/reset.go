package command

import (
	"flag"

	"github.com/gooolib/migration/migrate"
)

type ResetCommand struct {
	args      flag.FlagSet
	migration *migrate.Migration

	hard bool
}

func (c *ResetCommand) Exec() error {
	if c.hard {
		return c.migration.HardReset()
	}

	return c.migration.SoftReset()
}

func (c *ResetCommand) ParseArgs() error {
	hard := c.args.Lookup("hard")
	if hard != nil {
		c.hard = true
	}
	return nil
}
