package main

import "github.com/urfave/cli"

// NewApp return new application
func NewApp() *cli.App {
	app := cli.NewApp()
	app.Name = Name
	app.Usage = "Deploy and manage containerized applications like Heroku on AWS"
	app.Version = Version

	return app
}
