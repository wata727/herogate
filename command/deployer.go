package command

import (
	"github.com/urfave/cli"
	"github.com/wata727/herogate/herogate/deployer"
)

func DeployerCommand() cli.Command {
	return cli.Command{
		Name:  "deployer",
		Usage: "Manage Herogate deployer component.",
		Subcommands: []cli.Command{
			deployerLogsCommand(),
		},
	}
}

func deployerLogsCommand() cli.Command {
	return cli.Command{
		Name:   "logs",
		Usage:  "Display deployer logs (ECS Service events).",
		Flags:  deployerLogsFlags(),
		Action: deployer.Logs,
	}
}

func deployerLogsFlags() []cli.Flag {
	return []cli.Flag{
		cli.BoolFlag{
			Name:  "tail, t",
			Usage: "Continually display logs",
		},
	}
}
