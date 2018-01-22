package command

import (
	"github.com/urfave/cli"
	"github.com/wata727/herogate/herogate/builder"
)

func BuilderCommand() cli.Command {
	return cli.Command{
		Name:  "builder",
		Usage: "Manage Herogate builder component.",
		Subcommands: []cli.Command{
			builderLogsCommand(),
		},
	}
}

func builderLogsCommand() cli.Command {
	return cli.Command{
		Name:   "logs",
		Usage:  "Display builder logs (CodeBuild).",
		Flags:  builderLogsFlags(),
		Action: builder.Logs,
	}
}

func builderLogsFlags() []cli.Flag {
	return []cli.Flag{
		cli.BoolFlag{
			Name:  "tail, t",
			Usage: "Continually display logs",
		},
	}
}
