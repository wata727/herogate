package herogate

import (
	"bufio"
	"fmt"
	"io"
	"net/url"
	"os"
	"regexp"
	"strings"
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

type appsContext struct {
	app    *cli.App
	client iface.ClientInterface
}

// Apps returns your apps.
func Apps(ctx *cli.Context) {
	processApps(&appsContext{
		app: ctx.App,
		client: api.NewClient(&api.ClientOption{
			Region: "us-east-1", // NOTE: Currently, Fargate supported region is only `us-east-1`
		}),
	})
}

func processApps(ctx *appsContext) {
	apps := ctx.client.ListApps()

	fmt.Fprintln(ctx.app.Writer, "=== Apps")

	for _, app := range apps {
		fmt.Fprintln(ctx.app.Writer, app.Name)
	}

	fmt.Fprint(ctx.app.Writer, "\n")
}

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
	errorColor := color.New(color.FgRed)
	matched, err := regexp.MatchString(`^[a-z0-9][a-z-0-9_\-]+[a-z0-9]$`, ctx.name)
	if err != nil {
		return fmt.Errorf("%s    Failed to validate the application name", errorColor.Sprint("▸"))
	}
	if !matched {
		return fmt.Errorf("%s    The application name must match the pattern of `^[a-z0-9][a-z-0-9_\\-]+[a-z0-9]$`", errorColor.Sprint("▸"))
	}

	if _, err := ctx.client.GetApp(ctx.name); err == nil {
		return fmt.Errorf("%s    Name is already taken", errorColor.Sprint("▸"))
	}
	if ctx.client.StackExists(ctx.name) {
		return fmt.Errorf("%s    Cannot use already existing CloudFormation stack name", errorColor.Sprint("▸"))
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

type appsInfoContext struct {
	name   string
	app    *cli.App
	client iface.ClientInterface
}

// AppsInfo displays the application details.
// It includes platform version, container definiations, endpoint and repository URL.
func AppsInfo(ctx *cli.Context) error {
	_, name := detectAppFromRepo()
	if ctx.String("app") != "" {
		logrus.Debug("Override application name: " + ctx.String("app"))
		name = ctx.String("app")
	}
	if ctx.Args().First() != "" {
		logrus.Debug("Override application name: " + ctx.Args().First())
		name = ctx.Args().First()
	}
	if name == "" {
		return cli.NewExitError(
			fmt.Sprintf(
				"%s    No app specified.\n%s    USAGE: herogate apps:info APPNAME",
				color.New(color.FgRed).Sprint("▸"),
				color.New(color.FgRed).Sprint("▸"),
			),
			1,
		)
	}

	return processAppsInfo(&appsInfoContext{
		name: name,
		app:  ctx.App,
		client: api.NewClient(&api.ClientOption{
			Region: "us-east-1", // NOTE: Currently, Fargate supported region is only `us-east-1`
		}),
	})
}

func processAppsInfo(ctx *appsInfoContext) error {
	app, err := ctx.client.GetAppInfo(ctx.name)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("%s    Couldn't find that app.", color.New(color.FgRed).Sprint("▸")), 1)
	}

	fmt.Fprintln(ctx.app.Writer, fmt.Sprintf("=== %s", app.Name))
	for i, container := range app.Containers {
		if i == 0 {
			fmt.Fprintln(ctx.app.Writer, fmt.Sprintf("Containers:       %s: %d", container.Name, container.Count))
		} else {
			fmt.Fprintln(ctx.app.Writer, fmt.Sprintf("                  %s: %d", container.Name, container.Count))
		}
	}
	if app.Endpoint != "" {
		fmt.Fprintln(ctx.app.Writer, fmt.Sprintf("Web URL:          %s", app.Endpoint))
	}
	if app.Repository != "" {
		fmt.Fprintln(ctx.app.Writer, fmt.Sprintf("Git URL:          %s", app.Repository))
	}
	fmt.Fprintln(ctx.app.Writer, fmt.Sprintf("Status:           %s", app.Status))
	fmt.Fprintln(ctx.app.Writer, fmt.Sprintf("Region:           %s", app.Region))
	fmt.Fprintln(ctx.app.Writer, fmt.Sprintf("Platform Version: %s", app.PlatformVersion))

	return nil
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
		return cli.NewExitError(fmt.Sprintf("%s    Missing require flag `-a`, You must specify an application name", color.New(color.FgRed).Sprint("▸")), 1)
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
		return cli.NewExitError(fmt.Sprintf("%s    Couldn't find that app.", color.New(color.FgRed).Sprint("▸")), 1)
	}
	if app.Endpoint == "" {
		return cli.NewExitError(fmt.Sprintf("%s    Couldn't open that app because it doesn't have an endpoint.", color.New(color.FgRed).Sprint("▸")), 1)
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
		return cli.NewExitError(fmt.Sprintf("%s    Opening the app error: ", color.New(color.FgRed).Sprint("▸"))+err.Error(), 1)
	}
	return nil
}

type appsDestroyContext struct {
	name    string
	app     *cli.App
	confirm string
	client  iface.ClientInterface
}

var stdin io.Reader = os.Stdin

// AppsDestroy destroys the application.
// After that, it removes herogate remote from local git repository.
func AppsDestroy(ctx *cli.Context) error {
	_, name := detectAppFromRepo()
	if ctx.String("app") != "" {
		logrus.Debug("Override application name: " + ctx.String("app"))
		name = ctx.String("app")
	}
	if ctx.Args().First() != "" {
		logrus.Debug("Override application name: " + ctx.Args().First())
		name = ctx.Args().First()
	}
	if name == "" {
		return cli.NewExitError(
			fmt.Sprintf(
				"%s    No app specified.\n%s    USAGE: herogate apps:destroy APPNAME",
				color.New(color.FgRed).Sprint("▸"),
				color.New(color.FgRed).Sprint("▸"),
			),
			1,
		)
	}

	return processAppsDestroy(&appsDestroyContext{
		name:    name,
		app:     ctx.App,
		confirm: ctx.String("confirm"),
		client: api.NewClient(&api.ClientOption{
			Region: "us-east-1", // NOTE: Currently, Fargate supported region is only `us-east-1`
		}),
	})
}

func processAppsDestroy(ctx *appsDestroyContext) error {
	_, err := ctx.client.GetApp(ctx.name)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("%s    Couldn't find that app.", color.New(color.FgRed).Sprint("▸")), 1)
	}

	if err = confirmAppDeletion(ctx); err != nil {
		return err
	}

	ch := make(chan error, 1)
	go func() {
		ch <- ctx.client.DestroyApp(ctx.name)
	}()

	r, w := io.Pipe()
	go func() {
		defer w.Close()
		fmt.Fprintf(w, "Destroying %s... %d%%\r", color.New(color.FgMagenta).Sprintf("⬢ %s", ctx.name), 0)
		waitDeletionAndWriteProgress(ctx, w, ch)
	}()

	io.Copy(ctx.app.Writer, r)

	return nil
}

func confirmAppDeletion(ctx *appsDestroyContext) error {
	warnColor := color.New(color.FgYellow)
	errorColor := color.New(color.FgRed)
	appColor := color.New(color.FgMagenta)

	name := ctx.confirm

	if name == "" {
		fmt.Fprintf(ctx.app.Writer, "%s    WARNING: This will delete %s\n", warnColor.Sprint("▸"), appColor.Sprintf("⬢ %s", ctx.name))
		fmt.Fprintf(ctx.app.Writer, "%s    To proceed, type %s or re-run this command with\n", warnColor.Sprint("▸"), errorColor.Sprint(ctx.name))
		fmt.Fprintf(ctx.app.Writer, "%s    %s\n\n", warnColor.Sprint("▸"), errorColor.Sprintf("--confirm %s", ctx.name))
		fmt.Fprint(ctx.app.Writer, "> ")

		scanner := bufio.NewScanner(stdin)
		scanner.Scan()
		name = strings.TrimSpace(scanner.Text())
	}

	if name != ctx.name {
		return cli.NewExitError(fmt.Sprintf("%s    Confirmation did not match %s. Aborted.", errorColor.Sprint("▸"), errorColor.Sprint(ctx.name)), 1)
	}

	return nil
}

func waitDeletionAndWriteProgress(ctx *appsDestroyContext, w io.Writer, ch chan error) {
	select {
	case err := <-ch:
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"appName": ctx.name,
			}).Fatal("Failed to destroy the app: " + err.Error())
		}
		deleteLocalRepository()
		fmt.Fprintf(w, "Destroying %s... done\n", color.New(color.FgMagenta).Sprintf("⬢ %s", ctx.name))
	default:
		time.Sleep(progressCheckInterval)
		percent := ctx.client.GetAppDeletionProgress(ctx.name)
		fmt.Fprintf(w, "Destroying %s... %d%%\r", color.New(color.FgMagenta).Sprintf("⬢ %s", ctx.name), percent)
		waitDeletionAndWriteProgress(ctx, w, ch)
	}
}

func deleteLocalRepository() {
	repo, err := git.PlainOpen(".")
	if err != nil {
		logrus.Debug("Failed to open local Git repository: " + err.Error())
		return
	}

	err = repo.DeleteRemote("herogate")
	if err != nil {
		logrus.Debug("Failed to delete remote: " + err.Error())
		return
	}
}
