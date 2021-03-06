package command

import (
	"github.com/urfave/cli"
	"github.com/wata727/herogate/herogate"
)

// LogsCommand is command for retrieving Herogate logs.
func LogsCommand() cli.Command {
	return cli.Command{
		Name:   "logs",
		Usage:  "display recent log output",
		Flags:  append(sharedFlags(), logsFlags()...),
		Action: herogate.Logs,
	}
}

func logsFlags() []cli.Flag {
	return []cli.Flag{
		cli.IntFlag{
			Name:  "num, n",
			Value: 100,
			Usage: "number of lines to display",
		},
		cli.StringFlag{
			Name:  "ps, p",
			Usage: "process to limit filter by",
		},
		cli.StringFlag{
			Name:  "source, s",
			Usage: "log source to limit filter by",
		},
		cli.BoolFlag{
			Name:  "tail, t",
			Usage: "continually stream logs",
		},
	}
}
