package herogate

import (
	"fmt"
	"io"
	"sort"
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

	fmt.Fprintln(ctx.app.Writer, fmt.Sprintf("=== %s Config Vars", ctx.name))
	putsEnvVars(envVars, ctx.app.Writer)

	return nil
}

func putsEnvVars(envVars map[string]string, writer io.Writer) {
	var rightLength int
	envList := []map[string]string{}
	for key, value := range envVars {
		if rightLength < len(key) {
			rightLength = len(key)
		}
		envList = append(envList, map[string]string{"Name": key, "Value": value})
	}

	// Sort alphabetically
	sort.Slice(envList, func(i, j int) bool {
		return envList[i]["Name"] < envList[j]["Name"]
	})

	for _, env := range envList {
		// 2 = colon + space
		str := gopad.Right(fmt.Sprintf("%s:", env["Name"]), rightLength+2)
		fmt.Fprintln(writer, strings.Replace(str, env["Name"], color.New(color.FgGreen).Sprint(env["Name"]), 1)+env["Value"])
	}
}

type configGetContext struct {
	name   string
	env    string
	app    *cli.App
	client iface.ClientInterface
}

// ConfigGet displays an environment variable of the application container.
func ConfigGet(ctx *cli.Context) error {
	_, name := detectAppFromRepo()
	if ctx.String("app") != "" {
		logrus.Debug("Override application name: " + ctx.String("app"))
		name = ctx.String("app")
	}
	if name == "" {
		return cli.NewExitError(fmt.Sprintf("%s    Missing require flag `-a`, You must specify an application name", color.New(color.FgRed).Sprint("▸")), 1)
	}

	env := ctx.Args().First()
	if env == "" {
		return cli.NewExitError(fmt.Sprintf("%s    Missing require argument, You must specify an environment variable name", color.New(color.FgRed).Sprint("▸")), 1)
	}

	return processConfigGet(&configGetContext{
		name: name,
		env:  env,
		app:  ctx.App,
		client: api.NewClient(&api.ClientOption{
			Region: "us-east-1", // NOTE: Currently, Fargate supported region is only `us-east-1`
		}),
	})
}

func processConfigGet(ctx *configGetContext) error {
	envVars, err := ctx.client.DescribeEnvVars(ctx.name)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("%s    Couldn't find that app.", color.New(color.FgRed).Sprint("▸")), 1)
	}

	var env string
	for key, value := range envVars {
		if key == ctx.env {
			env = value
		}
	}

	fmt.Fprintln(ctx.app.Writer, env)

	return nil
}

type configSetContext struct {
	name   string
	args   []string
	app    *cli.App
	client iface.ClientInterface
}

// ConfigSet injects environment variables to application containers.
func ConfigSet(ctx *cli.Context) error {
	_, name := detectAppFromRepo()
	if ctx.String("app") != "" {
		logrus.Debug("Override application name: " + ctx.String("app"))
		name = ctx.String("app")
	}
	if name == "" {
		return cli.NewExitError(fmt.Sprintf("%s    Missing require flag `-a`, You must specify an application name", color.New(color.FgRed).Sprint("▸")), 1)
	}
	if !ctx.Args().Present() {
		return cli.NewExitError(fmt.Sprintf("%s    Missing require argument, You must specify key value pairs of environment variables", color.New(color.FgRed).Sprint("▸")), 1)
	}

	return processConfigSet(&configSetContext{
		name: name,
		args: ctx.Args(),
		app:  ctx.App,
		client: api.NewClient(&api.ClientOption{
			Region: "us-east-1", // NOTE: Currently, Fargate supported region is only `us-east-1`
		}),
	})
}

func processConfigSet(ctx *configSetContext) error {
	envVars := map[string]string{}
	envList := []string{}
	for _, arg := range ctx.args {
		env := strings.SplitN(arg, "=", 2)
		if len(env) == 1 {
			return cli.NewExitError(
				fmt.Sprintf(
					"%s    %s is invalid. Must be in the format %s.",
					color.New(color.FgRed).Sprint("▸"),
					color.New(color.FgCyan).Sprint(env[0]),
					color.New(color.FgCyan).Sprint("FOO=bar"),
				),
				1)
		}
		envVars[env[0]] = env[1]
		envList = append(envList, color.New(color.FgGreen).Sprint(env[0]))
	}

	_, err := ctx.client.GetApp(ctx.name)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("%s    Couldn't find that app.", color.New(color.FgRed).Sprint("▸")), 1)
	}

	appStr := color.New(color.FgMagenta).Sprintf("⬢ %s", ctx.name)
	fmt.Fprintf(ctx.app.Writer, "Setting %s and restarting %s...\r", strings.Join(envList, ", "), appStr)

	err = ctx.client.SetEnvVars(ctx.name, envVars)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"appName": ctx.name,
		}).Fatal("Failed to set environment variables: " + err.Error())
	}

	fmt.Fprintf(ctx.app.Writer, "Setting %s and restarting %s... done\n", strings.Join(envList, ", "), appStr)
	putsEnvVars(envVars, ctx.app.Writer)

	return nil
}
