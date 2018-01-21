package main

import (
	"github.com/urfave/cli"
	"github.com/wata727/herogate/command"
)

// NewApp return new application
func NewApp() *cli.App {
	app := cli.NewApp()
	app.Name = Name
	app.Usage = "Deploy and manage containerized applications like Heroku on AWS"
	app.Version = Version

	app.Commands = []cli.Command{
		command.BuilderCommand(),
		command.DeployerCommand(),
	}

	return app
}
