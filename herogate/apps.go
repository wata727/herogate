package herogate

import (
	"fmt"
	"io"
	"net/url"
	"regexp"
	"time"

	haikunator "github.com/Atrox/haikunatorgo"
	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
	"github.com/skratchdot/open-golang/open"
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

var progressCheckInterval = 10 * time.Second

// AppsCreate creates a new app with application name provided from CLI.
// If application name is not provided, This action creates Heroku-like
// random application name.
func AppsCreate(ctx *cli.Context) error {
	name := ctx.Args().First()
	if name == "" {
		haikunator := haikunator.New()
		name = haikunator.Haikunate()
	}

	return processAppsCreate(&appsCreateContext{
		name: name,
		app:  ctx.App,
		client: api.NewClient(&api.ClientOption{
			Region: "us-east-1", // NOTE: Currently, Fargate supported region is only `us-east-1`
		}),
	})
}

func processAppsCreate(ctx *appsCreateContext) error {
	if err := validateAppName(ctx); err != nil {
		return cli.NewExitError(err.Error(), 1)
	}

	ch := make(chan appsCreateOutput, 1)
	go func() {
		app := ctx.client.CreateApp(ctx.name)
		ch <- appsCreateOutput{
			repository: app.Repository,
			endpoint:   app.Endpoint,
		}
	}()

	r, w := io.Pipe()
	go func() {
		defer w.Close()
		fmt.Fprintf(w, "Creating app... %d%%\r", 0)
		waitCreationAndWriteProgress(ctx, w, ch)
	}()

	io.Copy(ctx.app.Writer, r)

	return nil
}

func validateAppName(ctx *appsCreateContext) error {
	matched, err := regexp.MatchString(`^[a-z0-9][a-z-0-9_\-]+[a-z0-9]$`, ctx.name)
	if err != nil {
		return fmt.Errorf("ERROR: Failed to validate app name: %s", ctx.name)
	}
	if !matched {
		return fmt.Errorf("ERROR: The application name must match the pattern of `^[a-z0-9][a-z-0-9_\\-]+[a-z0-9]$`: %s", ctx.name)
	}

	if _, err := ctx.client.GetApp(ctx.name); err == nil {
		return fmt.Errorf("ERROR: The application name already exists")
	}

	return nil
}

func waitCreationAndWriteProgress(ctx *appsCreateContext, w io.Writer, ch chan appsCreateOutput) {
	select {
	case v := <-ch:
		writeCreatedRepository(v.repository)
		writeCreationResult(ctx.name, v, w)
	default:
		time.Sleep(progressCheckInterval)
		percent := ctx.client.GetAppCreationProgress(ctx.name)
		fmt.Fprintf(w, "Creating app... %d%%\r", percent)
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

type appsOpenContext struct {
	name   string
	path   string
	client iface.ClientInterface
}

var openBrowser = open.Run

// AppsOpen opens the application endpoint on default browser.
// If you pass a path, opens the endpoint with the path.
func AppsOpen(ctx *cli.Context) error {
	path := ctx.Args().First()
	_, name := detectAppFromRepo()
	if ctx.String("app") != "" {
		logrus.Debug("Override application name: " + ctx.String("app"))
		name = ctx.String("app")
	}
	if name == "" {
		return cli.NewExitError("ERROR: Missing require flag `-a`, You must specify application name", 1)
	}

	return processAppsOpen(&appsOpenContext{
		name: name,
		path: path,
		client: api.NewClient(&api.ClientOption{
			Region: "us-east-1", // NOTE: Currently, Fargate supported region is only `us-east-1`
		}),
	})
}

func processAppsOpen(ctx *appsOpenContext) error {
	app, err := ctx.client.GetApp(ctx.name)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("ERROR: Application not found: %s", ctx.name), 1)
	}

	endpoint, err := url.Parse(app.Endpoint)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"appName":  ctx.name,
			"endpoint": app.Endpoint,
		}).Fatal("Failed to parse endpoint URL: " + err.Error())
	}
	if ctx.path != "" {
		endpoint, err = endpoint.Parse(ctx.path)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"appName":  ctx.name,
				"endpoint": app.Endpoint,
				"path":     ctx.path,
			}).Fatal("Failed to parse endpoint path: " + err.Error())
		}
	}

	err = openBrowser(endpoint.String())
	if err != nil {
		return cli.NewExitError("ERROR: Opening the app error: "+err.Error(), 1)
	}
	return nil
}
