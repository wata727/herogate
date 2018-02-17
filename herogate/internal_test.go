package herogate

import (
	"bytes"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/urfave/cli"
	"github.com/wata727/herogate/mock"
)

func TestProcessInternalGenerateTemplate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	app := cli.NewApp()
	writer := new(bytes.Buffer)
	app.Writer = writer

	client := mock.NewMockClientInterface(ctrl)
	// Expect to get template
	client.EXPECT().GetTemplate("bold-art-6993").Return(`
AWSTemplateFormatVersion: 2010-09-09
Description: Herogate Platform Template v1.0
Resources:
  HerogateApplicationContainer:
    Type: "AWS::ECS::TaskDefinition"
    Properties:
      ContainerDefinitions:
        - Name: web
          Image: "httpd:2.4"
`)

	processInternalGenerateTemplate(&internalGenerateTemplateContext{
		name:   "bold-art-6993",
		image:  "wata727/ecs-demo-php-simple-app",
		app:    app,
		client: client,
	})

	expected := `AWSTemplateFormatVersion: 2010-09-09
Description: Herogate Platform Template v1.0
Resources:
  HerogateApplicationContainer:
    Properties:
      ContainerDefinitions:
      - Image: wata727/ecs-demo-php-simple-app
        Name: web
    Type: AWS::ECS::TaskDefinition

`

	if writer.String() != expected {
		t.Fatalf("Expected template is `%s`, but get `%s`", expected, writer.String())
	}
}
