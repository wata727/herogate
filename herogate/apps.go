package herogate

import (
	"fmt"
	"regexp"
	"time"

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

type appsCreateOutput struct {
	repository string
	endpoint   string
}

// AppsCreate creates a new app with application name provided from CLI.
// If application name is not provided, This action creates Heroku-like
// random application name.
func AppsCreate(ctx *cli.Context) error {
	name := ctx.Args().First()
	if name == "" {
		haikunator := haikunator.New()
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
	ch := make(chan appsCreateOutput, 1)
	go func() {
		repository, endpoint := ctx.client.CreateApp(ctx.name)
		ch <- appsCreateOutput{
			repository: repository,
			endpoint:   endpoint,
		}
	}()
	fmt.Fprintln(ctx.app.Writer, "Creating app...")
	waitCreationAndWriteProgress(ctx, ch)
}

func waitCreationAndWriteProgress(ctx *appsCreateContext, ch chan appsCreateOutput) {
	select {
	case v := <-ch:
		fmt.Printf("repository: %s\n", v.repository)
		fmt.Printf("endpoint: %s\n", v.endpoint)
	default:
		time.Sleep(10 * time.Second)
		percent := ctx.client.GetAppCreationProgress(ctx.name)
		// TODO: More rich progress
		fmt.Fprintln(ctx.app.Writer, fmt.Sprintf("%d%% Completed", percent))
		waitCreationAndWriteProgress(ctx, ch)
	}
}
