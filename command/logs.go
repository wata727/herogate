package command

import (
	"github.com/urfave/cli"
	"github.com/wata727/herogate/logs"
)

func LogsCommand() cli.Command {
	return cli.Command{
		Name:   "logs",
		Usage:  "Display application or system logs.",
		Flags:  logsFlags(),
		Action: logs.Logs,
	}
}

func logsFlags() []cli.Flag {
	return []cli.Flag{
		cli.BoolFlag{
			Name:  "system",
			Usage: "Display system logs.",
		},
	}
}
