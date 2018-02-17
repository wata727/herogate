package command

import (
	"github.com/urfave/cli"
	"github.com/wata727/herogate/herogate"
)

// InternalCommand is a command for internal function such as builder.
func InternalCommand() cli.Command {
	return cli.Command{
		Name:   "internal",
		Hidden: true,
		Subcommands: []cli.Command{
			generateTemplateCommand(),
		},
	}
}

func generateTemplateCommand() cli.Command {
	return cli.Command{
		Name:   "generate-template",
		Action: herogate.InternalGenerateTemplate,
	}
}
