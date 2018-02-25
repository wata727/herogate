package herogate

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"

	"github.com/fatih/color"
	"github.com/golang/mock/gomock"
	"github.com/urfave/cli"
	"github.com/wata727/herogate/api"
	"github.com/wata727/herogate/api/objects"
	"github.com/wata727/herogate/mock"
)

func TestProcessApps(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := mock.NewMockClientInterface(ctrl)
	// Expect to list applications
	client.EXPECT().ListApps().Return([]*objects.App{
		{
			Name:       "young-eyrie-24091",
			Status:     "CREATE_COMPLETE",
			Repository: "ssh://git-codecommit.us-east-1.amazonaws.com/v1/repos/young-eyrie-24091",
			Endpoint:   "http://young-eyrie-24091-123456789.us-east-1.elb.amazonaws.com",
		},
		{
			Name:       "proud-lab-1661",
			Status:     "CREATE_COMPLETE",
			Repository: "ssh://git-codecommit.us-east-1.amazonaws.com/v1/repos/proud-lab-1661",
			Endpoint:   "http://proud-lab-1661-123456789.us-east-1.elb.amazonaws.com",
		},
	})

	app := cli.NewApp()
	writer := new(bytes.Buffer)
	app.Writer = writer

	processApps(&appsContext{
		app:    app,
		client: client,
	})

	expectedHeader := "=== Apps"
	expectedApp1 := "young-eyrie-24091"
	expectedApp2 := "proud-lab-1661"

	if !strings.Contains(writer.String(), expectedHeader) {
		t.Fatalf("Expected application outputs are not contained:\nExpected: %s\nActual: %s", expectedHeader, writer.String())
	}
	if !strings.Contains(writer.String(), expectedApp1) {
		t.Fatalf("Expected application outputs are not contained:\nExpected: %s\nActual: %s", expectedApp1, writer.String())
	}
	if !strings.Contains(writer.String(), expectedApp2) {
		t.Fatalf("Expected application outputs are not contained:\nExpected: %s\nActual: %s", expectedApp2, writer.String())
	}
}

func TestProcessAppsCreate(t *testing.T) {
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatal("Failed to get current directory: " + err.Error())
	}
	defer os.Chdir(currentDir)

	dir, err := ioutil.TempDir("", "repository")
	if err != nil {
		t.Fatal("Failed to create tempdir: " + err.Error())
	}
	defer os.RemoveAll(dir)

	repo, err := git.PlainInit(dir, false)
	if err != nil {
		t.Fatal("Failed to init git reporisoty: " + err.Error())
	}

	err = os.Chdir(dir)
	if err != nil {
		t.Fatal("Failed to change directory: " + err.Error())
	}

	// Wait only 1 second
	progressCheckInterval = 1 * time.Second
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	app := cli.NewApp()
	writer := new(bytes.Buffer)
	app.Writer = writer

	client := mock.NewMockClientInterface(ctrl)
	// Expect to get application
	client.EXPECT().GetApp("young-eyrie-24091").Return(nil, errors.New("Stack not found"))
	// Expect to check stack
	client.EXPECT().StackExists("young-eyrie-24091").Return(false)
	// Expect to create application
	client.EXPECT().CreateApp("young-eyrie-24091").Return(&objects.App{
		Name:       "young-eyrie-24091",
		Status:     "CREATE_COMPLETE",
		Repository: "ssh://git-codecommit.us-east-1.amazonaws.com/v1/repos/young-eyrie-24091",
		Endpoint:   "http://young-eyrie-24091-123456789.us-east-1.elb.amazonaws.com",
	})
	// Allow to get progress rate
	client.EXPECT().GetAppCreationProgress("young-eyrie-24091").Return(100).AnyTimes()

	err = processAppsCreate(&appsCreateContext{
		name:   "young-eyrie-24091",
		app:    app,
		client: client,
	})
	if err != nil {
		t.Fatalf("Expected error is nil, but get `%s`", err.Error())
	}

	appColor := color.New(color.FgMagenta)
	expectedApp := fmt.Sprint("Creating app... done, " + appColor.Sprint("⬢ young-eyrie-24091"))
	endpointColor := color.New(color.FgCyan)
	repositoryColor := color.New(color.FgGreen)
	expectedResult := fmt.Sprintf(
		"%s | %s",
		endpointColor.Sprint("http://young-eyrie-24091-123456789.us-east-1.elb.amazonaws.com"),
		repositoryColor.Sprint("ssh://git-codecommit.us-east-1.amazonaws.com/v1/repos/young-eyrie-24091"),
	)

	if !strings.Contains(writer.String(), expectedApp) {
		t.Fatalf("Expected application outputs are not contained:\nExpected: %s\nActual: %s", expectedApp, writer.String())
	}
	if !strings.Contains(writer.String(), expectedResult) {
		t.Fatalf("Expected result are not contained:\nExpected: %s\nActual: %s", expectedResult, writer.String())
	}

	remote, err := repo.Remote("herogate")
	if err != nil {
		t.Fatal("Failed to load remote config: " + err.Error())
	}
	if len(remote.Config().URLs) == 0 {
		t.Fatal("Expected count of repository URLs is not 0, but this is 0.")
	}
	if remote.Config().URLs[0] != "ssh://git-codecommit.us-east-1.amazonaws.com/v1/repos/young-eyrie-24091" {
		t.Fatalf("Expected repository remote URL is `ssh://git-codecommit.us-east-1.amazonaws.com/v1/repos/young-eyrie-24091`, but get `%s`", remote.Config().URLs[0])
	}
}

func TestProcessAppsCreate__invalidName(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	err := processAppsCreate(&appsCreateContext{
		name:   "YoungEyrie-24091",
		app:    cli.NewApp(),
		client: api.NewClient(&api.ClientOption{}),
	})

	expected := cli.NewExitError(fmt.Sprintf("%s    The application name must match the pattern of `^[a-z0-9][a-z-0-9_\\-]+[a-z0-9]$`", color.New(color.FgRed).Sprint("▸")), 1)
	if err.Error() != expected.Error() {
		t.Fatalf("Expected error is `%s`, but get `%s`", expected.Error(), err.Error())
	}
}

func TestProcessAppsCreate__duplicateName(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := mock.NewMockClientInterface(ctrl)
	client.EXPECT().GetApp("young-eyrie-24091").Return(&objects.App{}, nil)

	err := processAppsCreate(&appsCreateContext{
		name:   "young-eyrie-24091",
		app:    cli.NewApp(),
		client: client,
	})

	expected := cli.NewExitError(fmt.Sprintf("%s    Name is already taken", color.New(color.FgRed).Sprint("▸")), 1)
	if err.Error() != expected.Error() {
		t.Fatalf("Expected error is `%s`, but get `%s`", expected.Error(), err.Error())
	}
}

func TestProcessAppsCreate__stackExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := mock.NewMockClientInterface(ctrl)
	// Expect to get application
	client.EXPECT().GetApp("young-eyrie-24091").Return(nil, errors.New("Stack not found"))
	// Expect to check stack
	client.EXPECT().StackExists("young-eyrie-24091").Return(true)

	err := processAppsCreate(&appsCreateContext{
		name:   "young-eyrie-24091",
		app:    cli.NewApp(),
		client: client,
	})

	expected := cli.NewExitError(fmt.Sprintf("%s    Cannot use already existing CloudFormation stack name", color.New(color.FgRed).Sprint("▸")), 1)
	if err.Error() != expected.Error() {
		t.Fatalf("Expected error is `%s`, but get `%s`", expected.Error(), err.Error())
	}
}

func TestProcessAppsOpen(t *testing.T) {
	var called bool
	// Mock opening browser function
	openBrowser = func(url string) error {
		if url == "http://young-eyrie-24091-123456789.us-east-1.elb.amazonaws.com/" {
			called = true
		} else {
			t.Fatalf("Expected arugument is `http://young-eyrie-24091-123456789.us-east-1.elb.amazonaws.com/`, but get `%s`", url)
		}
		return nil
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := mock.NewMockClientInterface(ctrl)
	// Expect to get application
	client.EXPECT().GetApp("young-eyrie-24091").Return(&objects.App{
		Name:       "young-eyrie-24091",
		Status:     "CREATE_COMPLETE",
		Repository: "ssh://git-codecommit.us-east-1.amazonaws.com/v1/repos/young-eyrie-24091",
		Endpoint:   "http://young-eyrie-24091-123456789.us-east-1.elb.amazonaws.com/",
	}, nil)

	err := processAppsOpen(&appsOpenContext{
		name:   "young-eyrie-24091",
		path:   "",
		client: client,
	})
	if err != nil {
		t.Fatalf("Expected error is nil, but get `%s`", err.Error())
	}

	if !called {
		t.Fatal("Expected to open browser, but does not")
	}
}

func TestProcessAppsOpen__withPath(t *testing.T) {
	var called bool
	// Mock opening browser function
	openBrowser = func(url string) error {
		if url == "http://young-eyrie-24091-123456789.us-east-1.elb.amazonaws.com/foo/bar" {
			called = true
		} else {
			t.Fatalf("Expected arugument is `http://young-eyrie-24091-123456789.us-east-1.elb.amazonaws.com/foo/bar`, but get `%s`", url)
		}
		return nil
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := mock.NewMockClientInterface(ctrl)
	// Expect to get application
	client.EXPECT().GetApp("young-eyrie-24091").Return(&objects.App{
		Name:       "young-eyrie-24091",
		Status:     "CREATE_COMPLETE",
		Repository: "ssh://git-codecommit.us-east-1.amazonaws.com/v1/repos/young-eyrie-24091",
		Endpoint:   "http://young-eyrie-24091-123456789.us-east-1.elb.amazonaws.com/",
	}, nil)

	err := processAppsOpen(&appsOpenContext{
		name:   "young-eyrie-24091",
		path:   "foo/bar",
		client: client,
	})
	if err != nil {
		t.Fatalf("Expected error is nil, but get `%s`", err.Error())
	}

	if !called {
		t.Fatal("Expected to open browser, but does not")
	}
}

func TestProcessAppsOpen__invalidAppName(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := mock.NewMockClientInterface(ctrl)
	// Expect to get application
	client.EXPECT().GetApp("young-eyrie-24091").Return(nil, errors.New("Stack not found"))

	err := processAppsOpen(&appsOpenContext{
		name:   "young-eyrie-24091",
		path:   "",
		client: client,
	})

	expected := fmt.Sprintf("%s    Couldn't find that app.", color.New(color.FgRed).Sprint("▸"))
	if err.Error() != expected {
		t.Fatalf("Expected error is `%s`, but get `%s`", expected, err.Error())
	}
}

func TestProcessAppsOpen__createInProgress(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := mock.NewMockClientInterface(ctrl)
	// Expect to get application
	client.EXPECT().GetApp("young-eyrie-24091").Return(&objects.App{
		Name:   "young-eyrie-24091",
		Status: "CREATE_IN_PROGRESS",
	}, nil)

	err := processAppsOpen(&appsOpenContext{
		name:   "young-eyrie-24091",
		path:   "",
		client: client,
	})

	expected := fmt.Sprintf("%s    Couldn't open that app because it doesn't have an endpoint.", color.New(color.FgRed).Sprint("▸"))
	if err.Error() != expected {
		t.Fatalf("Expected error is `%s`, but get `%s`", expected, err.Error())
	}
}

func TestProcessAppsOpen__failedOpen(t *testing.T) {
	// Mock opening browser function
	openBrowser = func(url string) error {
		return errors.New("Unexpected error occurred")
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := mock.NewMockClientInterface(ctrl)
	// Expect to get application
	client.EXPECT().GetApp("young-eyrie-24091").Return(&objects.App{
		Name:       "young-eyrie-24091",
		Status:     "CREATE_COMPLETE",
		Repository: "ssh://git-codecommit.us-east-1.amazonaws.com/v1/repos/young-eyrie-24091",
		Endpoint:   "http://young-eyrie-24091-123456789.us-east-1.elb.amazonaws.com/",
	}, nil)

	err := processAppsOpen(&appsOpenContext{
		name:   "young-eyrie-24091",
		path:   "",
		client: client,
	})

	expected := fmt.Sprintf("%s    Opening the app error: Unexpected error occurred", color.New(color.FgRed).Sprint("▸"))
	if err.Error() != expected {
		t.Fatalf("Expected error is `%s`, but get `%s`", expected, err.Error())
	}
}

func TestProcessAppsDestroy(t *testing.T) {
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatal("Failed to get current directory: " + err.Error())
	}
	defer os.Chdir(currentDir)

	dir, err := ioutil.TempDir("", "TestProcessAppsDestroy")
	if err != nil {
		t.Fatal("Failed to create tempdir: " + err.Error())
	}
	defer os.RemoveAll(dir)

	repo, err := git.PlainInit(dir, false)
	if err != nil {
		t.Fatal("Failed to init git reporisoty: " + err.Error())
	}
	_, err = repo.CreateRemote(&config.RemoteConfig{
		Name: "herogate",
		URLs: []string{"ssh://git-codecommit.us-east-1.amazonaws.com/v1/repos/young-eyrie-24091"},
	})
	if err != nil {
		t.Fatal("Failed to create remote: " + err.Error())
	}

	err = os.Chdir(dir)
	if err != nil {
		t.Fatal("Failed to change directory: " + err.Error())
	}

	// Wait only 1 second
	progressCheckInterval = 1 * time.Second
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	app := cli.NewApp()
	writer := new(bytes.Buffer)
	app.Writer = writer

	client := mock.NewMockClientInterface(ctrl)
	// Expect to get application
	client.EXPECT().GetApp("young-eyrie-24091").Return(&objects.App{
		Name:       "young-eyrie-24091",
		Status:     "CREATE_COMPLETE",
		Repository: "ssh://git-codecommit.us-east-1.amazonaws.com/v1/repos/young-eyrie-24091",
		Endpoint:   "http://young-eyrie-24091-123456789.us-east-1.elb.amazonaws.com/",
	}, nil)
	// Expect to destroy application
	client.EXPECT().DestroyApp("young-eyrie-24091").Return(nil)
	// Allow to get progress rate
	client.EXPECT().GetAppDeletionProgress("young-eyrie-24091").Return(100).AnyTimes()

	err = processAppsDestroy(&appsDestroyContext{
		name:    "young-eyrie-24091",
		app:     app,
		confirm: "young-eyrie-24091",
		client:  client,
	})
	if err != nil {
		t.Fatalf("Expected error is nil, but get `%s`", err.Error())
	}

	expectedApp := fmt.Sprintf("Destroying %s... done\n", color.New(color.FgMagenta).Sprint("⬢ young-eyrie-24091"))

	if !strings.Contains(writer.String(), expectedApp) {
		t.Fatalf("Expected application outputs are not contained:\nExpected: %s\nActual: %s", expectedApp, writer.String())
	}

	remotes, err := repo.Remotes()
	if err != nil {
		t.Fatal("Failed to load remote config: " + err.Error())
	}
	for _, remote := range remotes {
		if remote.Config().Name == "herogate" {
			t.Fatal("Failed to delete remote")
		}
	}
}

func TestProcessAppsDestroy__notFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	app := cli.NewApp()
	writer := new(bytes.Buffer)
	app.Writer = writer

	client := mock.NewMockClientInterface(ctrl)
	// Expect to get application
	client.EXPECT().GetApp("young-eyrie-24091").Return(nil, errors.New("stack not found"))

	err := processAppsDestroy(&appsDestroyContext{
		name:    "young-eyrie-24091",
		app:     app,
		confirm: "young-eyrie-24091",
		client:  client,
	})

	expected := fmt.Sprintf("%s    Couldn't find that app.", color.New(color.FgRed).Sprint("▸"))
	if err.Error() != expected {
		t.Fatalf("Expected error is `%s`, but get `%s`", expected, err.Error())
	}
}

func TestProcessAppsDestroy__confirmationFailed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	app := cli.NewApp()
	writer := new(bytes.Buffer)
	app.Writer = writer

	client := mock.NewMockClientInterface(ctrl)
	// Expect to get application
	client.EXPECT().GetApp("young-eyrie-24091").Return(&objects.App{
		Name:       "young-eyrie-24091",
		Status:     "CREATE_COMPLETE",
		Repository: "ssh://git-codecommit.us-east-1.amazonaws.com/v1/repos/young-eyrie-24091",
		Endpoint:   "http://young-eyrie-24091-123456789.us-east-1.elb.amazonaws.com/",
	}, nil)

	err := processAppsDestroy(&appsDestroyContext{
		name:    "young-eyrie-24091",
		app:     app,
		confirm: "young-eyrie",
		client:  client,
	})

	errorColor := color.New(color.FgRed)
	expected := fmt.Sprintf("%s    Confirmation did not match %s. Aborted.", errorColor.Sprint("▸"), errorColor.Sprint("young-eyrie-24091"))
	if err.Error() != expected {
		t.Fatalf("Expected error is `%s`, but get `%s`", expected, err.Error())
	}
}

func TestProcessAppsDestroy__withConfirmation(t *testing.T) {
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatal("Failed to get current directory: " + err.Error())
	}
	defer os.Chdir(currentDir)

	dir, err := ioutil.TempDir("", "TestProcessAppsDestroy")
	if err != nil {
		t.Fatal("Failed to create tempdir: " + err.Error())
	}
	defer os.RemoveAll(dir)

	repo, err := git.PlainInit(dir, false)
	if err != nil {
		t.Fatal("Failed to init git reporisoty: " + err.Error())
	}
	_, err = repo.CreateRemote(&config.RemoteConfig{
		Name: "herogate",
		URLs: []string{"ssh://git-codecommit.us-east-1.amazonaws.com/v1/repos/young-eyrie-24091"},
	})
	if err != nil {
		t.Fatal("Failed to create remote: " + err.Error())
	}

	err = os.Chdir(dir)
	if err != nil {
		t.Fatal("Failed to change directory: " + err.Error())
	}

	// Wait only 1 second
	progressCheckInterval = 1 * time.Second
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	app := cli.NewApp()
	writer := new(bytes.Buffer)
	app.Writer = writer

	client := mock.NewMockClientInterface(ctrl)
	// Expect to get application
	client.EXPECT().GetApp("young-eyrie-24091").Return(&objects.App{
		Name:       "young-eyrie-24091",
		Status:     "CREATE_COMPLETE",
		Repository: "ssh://git-codecommit.us-east-1.amazonaws.com/v1/repos/young-eyrie-24091",
		Endpoint:   "http://young-eyrie-24091-123456789.us-east-1.elb.amazonaws.com/",
	}, nil)
	// Expect to destroy application
	client.EXPECT().DestroyApp("young-eyrie-24091").Return(nil)
	// Allow to get progress rate
	client.EXPECT().GetAppDeletionProgress("young-eyrie-24091").Return(100).AnyTimes()

	// Write app name to os.Stdin
	r, w := io.Pipe()
	defer w.Close()
	stdin = r
	go func() {
		fmt.Fprintln(w, "young-eyrie-24091")
	}()

	err = processAppsDestroy(&appsDestroyContext{
		name:    "young-eyrie-24091",
		app:     app,
		confirm: "",
		client:  client,
	})
	if err != nil {
		t.Fatalf("Expected error is nil, but get `%s`", err.Error())
	}

	expectedApp := fmt.Sprintf("Destroying %s... done\n", color.New(color.FgMagenta).Sprint("⬢ young-eyrie-24091"))

	if !strings.Contains(writer.String(), expectedApp) {
		t.Fatalf("Expected application outputs are not contained:\nExpected: %s\nActual: %s", expectedApp, writer.String())
	}

	remotes, err := repo.Remotes()
	if err != nil {
		t.Fatal("Failed to load remote config: " + err.Error())
	}
	for _, remote := range remotes {
		if remote.Config().Name == "herogate" {
			t.Fatal("Failed to delete remote")
		}
	}
}
