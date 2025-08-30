package command

import (
	"flag"
	"fmt"

	"github.com/gooolib/migration/migrate"
)

type Command struct {
	Type      string
	Executor  CommandExecutor
	migration *migrate.Migration
}

type CommandExecutor interface {
	Exec() error
	ParseArgs() error
}

const USAGE = "Usage: migrate <command> args...\nAvailable commands:\nup, down(rollback), rollback, reset, generate, status"

var availableCommands = map[string]func(*migrate.Migration, *flag.FlagSet) CommandExecutor{
	"up": func(m *migrate.Migration, args *flag.FlagSet) CommandExecutor {
		return &UpCommand{migration: m, args: args}
	},
	"down": func(m *migrate.Migration, args *flag.FlagSet) CommandExecutor {
		return &DownCommand{migration: m, args: args}
	},
	"rollback": func(m *migrate.Migration, args *flag.FlagSet) CommandExecutor {
		return &RollbackCommand{migration: m, args: args}
	},
	"generate": func(m *migrate.Migration, args *flag.FlagSet) CommandExecutor {
		return &GenerateCommand{migration: m, args: args}
	},
	"status": func(m *migrate.Migration, args *flag.FlagSet) CommandExecutor {
		return &StatusCommand{migration: m, args: args}
	},
}

func NewCommand(m *migrate.Migration) (*Command, error) {
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		return nil, fmt.Errorf(USAGE)
	}
	cmdType := args[0]
	executorFactory, ok := availableCommands[cmdType]
	if !ok {
		return nil, fmt.Errorf("unknown command: %s\n%s", cmdType, USAGE)
	}

	argsFlagSet := flag.NewFlagSet(cmdType, flag.ExitOnError)
	if err := argsFlagSet.Parse(args[1:]); err != nil {
		return nil, err
	}

	executor := executorFactory(m, argsFlagSet)
	err := executor.ParseArgs()
	if err != nil {
		return nil, err
	}

	return &Command{
		Type:      cmdType,
		Executor:  executor,
		migration: m,
	}, nil
}

func (c *Command) Exec() error {
	version := c.migration.GetCurrentVersion()
	if version == "" {
		version = "initial"
	}
	fmt.Println("")
	fmt.Println("Command:", c.Type)
	fmt.Println("Current version:", version)

	return c.Executor.Exec()
}
