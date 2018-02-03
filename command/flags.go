package command

import "github.com/urfave/cli"

func sharedFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:  "app, a",
			Usage: "application name",
		},
	}
}
