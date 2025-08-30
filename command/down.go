package command

import (
	"flag"
	"fmt"

	"github.com/gooolib/migration/internal/migrate"
)

type DownCommand struct {
	Version   string
	DownAll   bool
	args      *flag.FlagSet
	migration *migrate.Migration
}

func (c *DownCommand) Exec() error {
	if c.DownAll {
		return c.migration.DownAll()
	}

	if c.Version != "" {
		file := c.migration.FindFileByVersion(c.Version, "down")
		if file == nil {
			return fmt.Errorf("migration file with version %s not found", c.Version)
		}
		return c.migration.RunSingleDown(*file)
	}

	return c.migration.Down()
}

func (c *DownCommand) ParseArgs() error {
	version := c.args.String("version", "", "migrate to specific version")
	if version != nil {
		c.Version = *version
	}
	downAll := c.args.Lookup("all")
	if downAll != nil {
		c.DownAll = true
	}
	return nil
}
