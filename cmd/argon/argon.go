package main

import (
	"os"

	"github.com/pkg/errors"
	cli "github.com/urfave/cli/v2"

	"github.com/peertechde/argon/pkg/logging"
)

var log = logging.Logger.WithField(logging.Sys, "main")

func main() {
	app := cli.NewApp()
	app.Name = "argon"
	app.Usage = ""
	app.Flags = []cli.Flag{}
	app.Commands = []*cli.Command{
		WriteCommand(),
		ReadCommand(),
		ListCommand(),
		StatCommand(),
		RemoveCommand(),
		RenameCommand(),
		ServerCommand(),
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatalln(err)
	}
}

func requiredFlag(clictx *cli.Context, flag string) error {
	return errors.Errorf("'%s %s' requires the '--%s' flag", clictx.App.HelpName,
		clictx.Command.Name, flag)
}
