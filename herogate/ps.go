package herogate

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"github.com/wata727/herogate/api"
	"github.com/wata727/herogate/api/iface"
)

type psContext struct {
	name   string
	app    *cli.App
	client iface.ClientInterface
}

// Ps returns containers in an app.
func Ps(ctx *cli.Context) error {
	_, name := detectAppFromRepo()
	if ctx.String("app") != "" {
		logrus.Debug("Override application name: " + ctx.String("app"))
		name = ctx.String("app")
	}
	if name == "" {
		return cli.NewExitError(fmt.Sprintf("%s    Missing require flag `-a`, You must specify an application name", color.New(color.FgRed).Sprint("▸")), 1)
	}

	return processPs(&psContext{
		name: name,
		app:  ctx.App,
		client: api.NewClient(&api.ClientOption{
			Region: "us-east-1", // NOTE: Currently, Fargate supported region is only `us-east-1`
		}),
	})
}

func processPs(ctx *psContext) error {
	app, err := ctx.client.GetAppInfo(ctx.name)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("%s    Couldn't find that app.", color.New(color.FgRed).Sprint("▸")), 1)
	}

	for _, container := range app.Containers {
		name := color.New(color.FgGreen).Sprint(container.Name)
		count := color.New(color.FgYellow).Sprint(container.Count)
		var command string
		if len(container.Command) == 0 {
			command = "No commands"
		} else {
			command = strings.Join(container.Command, " ")
		}
		fmt.Fprintln(ctx.app.Writer, fmt.Sprintf("=== %s (%s): %s", name, count, command))
		fmt.Fprint(ctx.app.Writer, "\n")
	}

	return nil
}
