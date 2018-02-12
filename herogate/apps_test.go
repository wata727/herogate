package herogate

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	git "gopkg.in/src-d/go-git.v4"

	"github.com/fatih/color"
	"github.com/golang/mock/gomock"
	"github.com/urfave/cli"
	"github.com/wata727/herogate/api"
	"github.com/wata727/herogate/api/objects"
	"github.com/wata727/herogate/mock"
)

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

	expected := cli.NewExitError("ERROR: The application name must match the pattern of `^[a-z0-9][a-z-0-9_\\-]+[a-z0-9]$`: YoungEyrie-24091", 1)
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

	expected := cli.NewExitError("ERROR: The application name already exists", 1)
	if err.Error() != expected.Error() {
		t.Fatalf("Expected error is `%s`, but get `%s`", expected.Error(), err.Error())
	}
}
