package command

import (
	"flag"
	"fmt"

	"github.com/gooolib/migration/internal/migrate"
)

type UpCommand struct {
	Version   string
	args      *flag.FlagSet
	migration *migrate.Migration
}

func (c *UpCommand) ParseArgs() error {
	version := c.args.String("version", "", "migrate to specific version")
	if version != nil {
		c.Version = *version
	}
	return nil
}

func (c *UpCommand) Exec() error {
	if c.Version != "" {
		file := c.migration.FindFileByVersion(c.Version, "up")
		if file == nil {
			return fmt.Errorf("migration file with version %s not found", c.Version)
		}
		return c.migration.RunSingleUp(*file)
	}

	return c.migration.Up()
}
