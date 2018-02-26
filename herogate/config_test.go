package herogate

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"github.com/fatih/color"
	"github.com/golang/mock/gomock"
	"github.com/urfave/cli"
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
%s:       production
%s:        production
%s: 011a60b8e222a55e0869e3dca9301a7736074189cb52782f1efd8b8a2e956fc44b25a6f2753f1662986c9519fbebdb7ebb4799becc75ac1a7faad0b55aee1b4b
`, color.New(color.FgGreen).Sprint("RAILS_ENV"), color.New(color.FgGreen).Sprint("RACK_ENV"), color.New(color.FgGreen).Sprint("SECRET_KEY_BASE"))

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

	expected := fmt.Sprintf("%s    Couldn't find that app.", color.New(color.FgRed).Sprint("▸"))
	if err.Error() != expected {
		t.Fatalf("Expected error is `%s`, but get `%s`", expected, err.Error())
	}
}
