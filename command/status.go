package command

import (
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/gooolib/migration/migrate"
)

type StatusCommand struct {
	args      *flag.FlagSet
	migration *migrate.Migration
}

func (c *StatusCommand) ParseArgs() error {
	return nil
}

func (c *StatusCommand) Exec() error {
	statuses, err := c.migration.Status()
	if err != nil {
		return err
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "")
	fmt.Fprintf(w, "|\tVersion\t|\tStatus\t|\n")
	fmt.Fprintln(w, "+\t=================\t+\t========\t+")
	for _, status := range statuses {
		fmt.Fprintln(w, fmt.Sprintf("|\t%s\t|\t%s\t|", status.Version, status.Status))
	}
	fmt.Fprintln(w, "+\t=================\t+\t========\t+")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "")
	w.Flush()

	return nil
}
