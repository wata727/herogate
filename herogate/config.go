package herogate

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/rhymond/gopad"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"github.com/wata727/herogate/api"
	"github.com/wata727/herogate/api/iface"
)

type configContext struct {
	name   string
	app    *cli.App
	client iface.ClientInterface
}

// Config displays environment variables of the application container.
func Config(ctx *cli.Context) error {
	_, name := detectAppFromRepo()
	if ctx.String("app") != "" {
		logrus.Debug("Override application name: " + ctx.String("app"))
		name = ctx.String("app")
	}
	if name == "" {
		return cli.NewExitError(fmt.Sprintf("%s    Missing require flag `-a`, You must specify an application name", color.New(color.FgRed).Sprint("▸")), 1)
	}

	return processConfig(&configContext{
		name: name,
		app:  ctx.App,
		client: api.NewClient(&api.ClientOption{
			Region: "us-east-1", // NOTE: Currently, Fargate supported region is only `us-east-1`
		}),
	})
}

func processConfig(ctx *configContext) error {
	envVars, err := ctx.client.DescribeEnvVars(ctx.name)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("%s    Couldn't find that app.", color.New(color.FgRed).Sprint("▸")), 1)
	}

	var rightLength int
	for key := range envVars {
		if rightLength < len(key) {
			rightLength = len(key)
		}
	}

	fmt.Fprintln(ctx.app.Writer, fmt.Sprintf("=== %s Config Vars", ctx.name))
	for key, value := range envVars {
		// 2 = colon + space
		str := gopad.Right(fmt.Sprintf("%s:", key), rightLength+2)
		fmt.Fprintln(ctx.app.Writer, strings.Replace(str, key, color.New(color.FgGreen).Sprint(key), 1)+value)
	}

	return nil
}
