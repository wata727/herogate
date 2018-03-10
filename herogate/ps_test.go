package herogate

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"github.com/fatih/color"
	"github.com/golang/mock/gomock"
	"github.com/urfave/cli"
	"github.com/wata727/herogate/api/objects"
	"github.com/wata727/herogate/mock"
)

func TestProcessPs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := mock.NewMockClientInterface(ctrl)
	// Expect to get application info
	client.EXPECT().GetAppInfo("young-eyrie-24091").Return(&objects.AppInfo{
		App: &objects.App{
			Name:            "young-eyrie-24091",
			Status:          "CREATE_COMPLETE",
			Repository:      "ssh://git-codecommit.us-east-1.amazonaws.com/v1/repos/young-eyrie-24091",
			Endpoint:        "http://young-eyrie-24091-123456789.us-east-1.elb.amazonaws.com",
			PlatformVersion: "1.0",
		},
		Containers: []*objects.Container{
			{
				Name:    "web",
				Count:   1,
				Command: []string{"bundle", "exec", "puma"},
			},
			{
				Name:    "worker",
				Count:   1,
				Command: []string{"bundle", "exec", "sidekiq"},
			},
		},
		Region: "us-east-1",
	}, nil)

	app := cli.NewApp()
	writer := new(bytes.Buffer)
	app.Writer = writer

	err := processPs(&psContext{
		name:   "young-eyrie-24091",
		app:    app,
		client: client,
	})
	if err != nil {
		t.Fatalf("Expected error is nil, but get `%s`", err.Error())
	}

	expected := fmt.Sprintf(`=== %s (%s): bundle exec puma

=== %s (%s): bundle exec sidekiq

`, color.New(color.FgGreen).Sprint("web"), color.New(color.FgYellow).Sprint(1), color.New(color.FgGreen).Sprint("worker"), color.New(color.FgYellow).Sprint(1))
	if writer.String() != expected {
		t.Fatalf("Expected to output is %s, but get `%s`", expected, writer.String())
	}
}

func TestProcessPs__invalidAppName(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := mock.NewMockClientInterface(ctrl)
	// Expect to get application
	client.EXPECT().GetAppInfo("young-eyrie-24091").Return(nil, errors.New("Stack not found"))

	err := processPs(&psContext{
		name:   "young-eyrie-24091",
		app:    cli.NewApp(),
		client: client,
	})

	expected := fmt.Sprintf("%s    Couldn't find that app.", color.New(color.FgRed).Sprint("▸"))
	if err.Error() != expected {
		t.Fatalf("Expected error is `%s`, but get `%s`", expected, err.Error())
	}
}
