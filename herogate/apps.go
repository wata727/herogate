package herogate

import (
	"fmt"
	"regexp"

	haikunator "github.com/Atrox/haikunatorgo"
	"github.com/urfave/cli"
	"github.com/wata727/herogate/api"
	"github.com/wata727/herogate/api/iface"
)

type appsCreateContext struct {
	name   string
	app    *cli.App
	client iface.ClientInterface
}

// AppsCreate creates a new app with application name provided from CLI.
// If application name is not provided, This action creates Heroku-like
// random application name.
func AppsCreate(ctx *cli.Context) error {
	name := ctx.Args().First()
	if name == "" {
		haikunator := haikunator.New()
		haikunator.TokenLength = 0
		name = haikunator.Haikunate()
	}

	if err := validateAppName(name); err != nil {
		return cli.NewExitError(err.Error(), 1)
	}

	processAppsCreate(&appsCreateContext{
		name:   name,
		app:    ctx.App,
		client: api.NewClient(&api.ClientOption{}),
	})

	return nil
}

func validateAppName(name string) error {
	matched, err := regexp.MatchString(`^[a-z0-9][a-z-0-9_\-]+[a-z0-9]$`, name)
	if err != nil {
		return fmt.Errorf("ERROR: Failed to validate app name: %s", name)
	}
	if !matched {
		return fmt.Errorf("ERROR: The application name must match the pattern of `^[a-z0-9][a-z-0-9_\\-]+[a-z0-9]$`: %s", name)
	}
	// TODO: validate duplicate app name
	return nil
}

func processAppsCreate(ctx *appsCreateContext) {
	fmt.Print(ctx.name)
}
