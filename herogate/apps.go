package herogate

import (
	"fmt"
	"io"
	"regexp"
	"time"

	haikunator "github.com/Atrox/haikunatorgo"
	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"github.com/wata727/herogate/api"
	"github.com/wata727/herogate/api/iface"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
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
			endpoint:   "http://" + endpoint,
		}
	}()

	r, w := io.Pipe()
	go func() {
		defer w.Close()
		waitCreationAndWriteProgress(ctx, w, ch)
	}()

	io.Copy(ctx.app.Writer, r)
}

func waitCreationAndWriteProgress(ctx *appsCreateContext, w io.Writer, ch chan appsCreateOutput) {
	select {
	case v := <-ch:
		writeCreatedRepository(v.repository)
		writeCreationResult(ctx.name, v, w)
	default:
		percent := ctx.client.GetAppCreationProgress(ctx.name)
		fmt.Fprintf(w, "Creating app... %d%%\r", percent)
		time.Sleep(10 * time.Second)
		waitCreationAndWriteProgress(ctx, w, ch)
	}
}

func writeCreatedRepository(repositoryURL string) {
	repo, err := git.PlainOpen(".")
	if err != nil {
		logrus.Debug("Failed to open local Git repository: " + err.Error())
		return
	}

	_, err = repo.CreateRemote(&config.RemoteConfig{
		Name: "herogate",
		URLs: []string{repositoryURL},
	})
	if err != nil {
		logrus.Debug("Failed to create remote: " + err.Error())
		return
	}
}

func writeCreationResult(appName string, v appsCreateOutput, w io.Writer) {
	appColor := color.New(color.FgMagenta)
	fmt.Fprintln(w, "Creating app... done, "+appColor.Sprintf("⬢ %s", appName))
	endpointColor := color.New(color.FgCyan)
	repositoryColor := color.New(color.FgGreen)
	fmt.Fprintf(w, "%s | %s\n", endpointColor.Sprint(v.endpoint), repositoryColor.Sprint(v.repository))
}
