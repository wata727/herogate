package herogate

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"github.com/fatih/color"
	"github.com/golang/mock/gomock"
	"github.com/urfave/cli"
	"github.com/wata727/herogate/api"
	"github.com/wata727/herogate/api/objects"
	"github.com/wata727/herogate/mock"
)

func TestProcessConfig(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := mock.NewMockClientInterface(ctrl)
	client.EXPECT().DescribeEnvVars("young-eyrie-24091").Return(map[string]string{
		"RAILS_ENV":       "production",
		"RACK_ENV":        "production",
		"SECRET_KEY_BASE": "011a60b8e222a55e0869e3dca9301a7736074189cb52782f1efd8b8a2e956fc44b25a6f2753f1662986c9519fbebdb7ebb4799becc75ac1a7faad0b55aee1b4b",
	}, nil)

	app := cli.NewApp()
	writer := new(bytes.Buffer)
	app.Writer = writer

	err := processConfig(&configContext{
		name:   "young-eyrie-24091",
		app:    app,
		client: client,
	})
	if err != nil {
		t.Fatalf("Expected error is nil, but get `%s`", err.Error())
	}

	expected := fmt.Sprintf(`=== young-eyrie-24091 Config Vars
%s:        production
%s:       production
%s: 011a60b8e222a55e0869e3dca9301a7736074189cb52782f1efd8b8a2e956fc44b25a6f2753f1662986c9519fbebdb7ebb4799becc75ac1a7faad0b55aee1b4b
`, color.New(color.FgGreen).Sprint("RACK_ENV"), color.New(color.FgGreen).Sprint("RAILS_ENV"), color.New(color.FgGreen).Sprint("SECRET_KEY_BASE"))

	if writer.String() != expected {
		t.Fatalf("Expected to output is %s, but get `%s`", expected, writer.String())
	}
}

func TestProcessConfig__invalidAppName(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := mock.NewMockClientInterface(ctrl)
	client.EXPECT().DescribeEnvVars("young-eyrie-24091").Return(map[string]string{}, errors.New("Stack not found"))

	err := processConfig(&configContext{
		name:   "young-eyrie-24091",
		app:    cli.NewApp(),
		client: client,
	})
	if err == nil {
		t.Fatal("Expected error is not nil, but get nil")
	}

	expected := fmt.Sprintf("%s    Couldn't find that app.", color.New(color.FgRed).Sprint("▸"))
	if err.Error() != expected {
		t.Fatalf("Expected error is `%s`, but get `%s`", expected, err.Error())
	}
}

func TestProcessConfigGet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := mock.NewMockClientInterface(ctrl)
	client.EXPECT().DescribeEnvVars("young-eyrie-24091").Return(map[string]string{
		"RAILS_ENV":       "production",
		"RACK_ENV":        "production",
		"SECRET_KEY_BASE": "011a60b8e222a55e0869e3dca9301a7736074189cb52782f1efd8b8a2e956fc44b25a6f2753f1662986c9519fbebdb7ebb4799becc75ac1a7faad0b55aee1b4b",
	}, nil)

	app := cli.NewApp()
	writer := new(bytes.Buffer)
	app.Writer = writer

	err := processConfigGet(&configGetContext{
		name:   "young-eyrie-24091",
		env:    "RAILS_ENV",
		app:    app,
		client: client,
	})
	if err != nil {
		t.Fatalf("Expected error is nil, but get `%s`", err.Error())
	}

	if writer.String() != "production\n" {
		t.Fatalf("Expected to output is %s, but get `%s`", "production\n", writer.String())
	}
}

func TestProcessConfigGet__invalidAppName(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := mock.NewMockClientInterface(ctrl)
	client.EXPECT().DescribeEnvVars("young-eyrie-24091").Return(map[string]string{}, errors.New("Stack not found"))

	err := processConfigGet(&configGetContext{
		name:   "young-eyrie-24091",
		env:    "RAILS_ENV",
		app:    cli.NewApp(),
		client: client,
	})
	if err == nil {
		t.Fatal("Expected error is not nil, but get nil")
	}

	expected := fmt.Sprintf("%s    Couldn't find that app.", color.New(color.FgRed).Sprint("▸"))
	if err.Error() != expected {
		t.Fatalf("Expected error is `%s`, but get `%s`", expected, err.Error())
	}
}

func TestProcessConfigSet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := mock.NewMockClientInterface(ctrl)
	// Expect to get application
	client.EXPECT().GetApp("young-eyrie-24091").Return(&objects.App{}, nil)
	// Expect to set environment variables
	client.EXPECT().SetEnvVars("young-eyrie-24091", map[string]string{
		"RAILS_ENV": "production",
		"RACK_ENV":  "production",
	}).Return(nil)

	app := cli.NewApp()
	writer := new(bytes.Buffer)
	app.Writer = writer

	err := processConfigSet(&configSetContext{
		name:   "young-eyrie-24091",
		args:   []string{"RAILS_ENV=production", "RACK_ENV=production"},
		app:    app,
		client: client,
	})
	if err != nil {
		t.Fatalf("Expected error is nil, but get `%s`", err.Error())
	}

	railsEnv := color.New(color.FgGreen).Sprint("RAILS_ENV")
	rackEnv := color.New(color.FgGreen).Sprint("RACK_ENV")
	appStr := color.New(color.FgMagenta).Sprint("⬢ young-eyrie-24091")
	expected := fmt.Sprintf("Setting %s, %s and restarting %s...\r", railsEnv, rackEnv, appStr)
	expected = expected + fmt.Sprintf(`Setting %s, %s and restarting %s... done
%s:  production
%s: production
`, railsEnv, rackEnv, appStr, rackEnv, railsEnv)

	if writer.String() != expected {
		t.Fatalf("Expected to output is `%s`, but get `%s`", expected, writer.String())
	}
}

func TestProcessConfigSet__invalidEnvFormat(t *testing.T) {
	err := processConfigSet(&configSetContext{
		name:   "young-eyrie-24091",
		args:   []string{"RAILS_ENV", "RACK_ENV"},
		app:    cli.NewApp(),
		client: api.NewClient(&api.ClientOption{}),
	})

	if err == nil {
		t.Fatal("Expected error is not nil, but get nil")
	}
	expected := fmt.Sprintf(
		"%s    %s is invalid. Must be in the format %s.",
		color.New(color.FgRed).Sprint("▸"),
		color.New(color.FgCyan).Sprint("RAILS_ENV"),
		color.New(color.FgCyan).Sprint("FOO=bar"),
	)
	if err.Error() != expected {
		t.Fatalf("Expected error is `%s`, but get `%s`", expected, err.Error())
	}
}

func TestProcessConfigSet__invalidAppName(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := mock.NewMockClientInterface(ctrl)
	// Expect to get application
	client.EXPECT().GetApp("young-eyrie-24091").Return(nil, errors.New("Stack not found"))

	err := processConfigSet(&configSetContext{
		name:   "young-eyrie-24091",
		args:   []string{"RAILS_ENV=production", "RACK_ENV=production"},
		app:    cli.NewApp(),
		client: client,
	})
	if err == nil {
		t.Fatal("Expected error is not nil, but get nil")
	}

	expected := fmt.Sprintf("%s    Couldn't find that app.", color.New(color.FgRed).Sprint("▸"))
	if err.Error() != expected {
		t.Fatalf("Expected error is `%s`, but get `%s`", expected, err.Error())
	}
}

func TestProcessConfigUnset(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := mock.NewMockClientInterface(ctrl)
	// Expect to get application
	client.EXPECT().GetApp("young-eyrie-24091").Return(&objects.App{}, nil)
	// Expect to set environment variables
	client.EXPECT().UnsetEnvVars("young-eyrie-24091", []string{
		"RAILS_ENV", "RACK_ENV",
	}).Return(nil)

	app := cli.NewApp()
	writer := new(bytes.Buffer)
	app.Writer = writer

	err := processConfigUnset(&configUnsetContext{
		name:    "young-eyrie-24091",
		envList: []string{"RAILS_ENV", "RACK_ENV"},
		app:     app,
		client:  client,
	})
	if err != nil {
		t.Fatalf("Expected error is nil, but get `%s`", err.Error())
	}

	railsEnv := color.New(color.FgGreen).Sprint("RAILS_ENV")
	rackEnv := color.New(color.FgGreen).Sprint("RACK_ENV")
	appStr := color.New(color.FgMagenta).Sprint("⬢ young-eyrie-24091")
	expected := fmt.Sprintf("Unsetting %s, %s and restarting %s...\r", railsEnv, rackEnv, appStr)
	expected = expected + fmt.Sprintf("Unsetting %s, %s and restarting %s... done\n", railsEnv, rackEnv, appStr)

	if writer.String() != expected {
		t.Fatalf("Expected to output is `%s`, but get `%s`", expected, writer.String())
	}
}

func TestProcessConfigUnset__invalidAppName(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := mock.NewMockClientInterface(ctrl)
	// Expect to get application
	client.EXPECT().GetApp("young-eyrie-24091").Return(nil, errors.New("Stack not found"))

	err := processConfigUnset(&configUnsetContext{
		name:    "young-eyrie-24091",
		envList: []string{"RAILS_ENV", "RACK_ENV"},
		app:     cli.NewApp(),
		client:  client,
	})
	if err == nil {
		t.Fatal("Expected error is not nil, but get nil")
	}

	expected := fmt.Sprintf("%s    Couldn't find that app.", color.New(color.FgRed).Sprint("▸"))
	if err.Error() != expected {
		t.Fatalf("Expected error is `%s`, but get `%s`", expected, err.Error())
	}
}
