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
Description: Herogate Platform Template v1.0
Resources:
  HerogateApplicationContainer:
    Properties:
      ContainerDefinitions:
      - Image: httpd:2.4
        Name: web
    Type: AWS::ECS::TaskDefinition
`)

	processInternalGenerateTemplate(&internalGenerateTemplateContext{
		name:     "bold-art-6993",
		image:    "myapp:0.1",
		procfile: "web: bundle exec puma\nworker: bundle exec sidekiq\n",
		app:      app,
		client:   client,
	})

	expected := `Description: Herogate Platform Template v1.0
Resources:
  HerogateApplicationContainer:
    Properties:
      ContainerDefinitions:
      - Name: web
        Image: myapp:0.1
        Command:
        - bundle
        - exec
        - puma
        Environment: []
        PortMappings:
        - ContainerPort: 80
        LogConfiguration:
          LogDriver: awslogs
          Options:
            awslogs-region:
              Ref: AWS::Region
            awslogs-group:
              Ref: HerogateApplicationContainerLogs
            awslogs-stream-prefix: web
      - Name: worker
        Image: myapp:0.1
        Command:
        - bundle
        - exec
        - sidekiq
        Environment: []
        PortMappings: []
        LogConfiguration:
          LogDriver: awslogs
          Options:
            awslogs-region:
              Ref: AWS::Region
            awslogs-group:
              Ref: HerogateApplicationContainerLogs
            awslogs-stream-prefix: worker
    Type: AWS::ECS::TaskDefinition

`

	if writer.String() != expected {
		t.Fatalf("Expected template is `%s`, but get `%s`", expected, writer.String())
	}
}

func TestProcessInternalGenerateTemplate__withEnvironment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	app := cli.NewApp()
	writer := new(bytes.Buffer)
	app.Writer = writer

	client := mock.NewMockClientInterface(ctrl)
	// Expect to get template
	client.EXPECT().GetTemplate("bold-art-6993").Return(`
Description: Herogate Platform Template v1.0
Resources:
  HerogateApplicationContainer:
    Properties:
      ContainerDefinitions:
      - Image: httpd:2.4
        Name: web
        Environment:
        - Name: RAILS_ENV
          Value: production
        - Name: RACK_ENV
          Value: production
    Type: AWS::ECS::TaskDefinition
`)

	processInternalGenerateTemplate(&internalGenerateTemplateContext{
		name:     "bold-art-6993",
		image:    "myapp:0.1",
		procfile: "web: bundle exec puma\nworker: bundle exec sidekiq\n",
		app:      app,
		client:   client,
	})

	expected := `Description: Herogate Platform Template v1.0
Resources:
  HerogateApplicationContainer:
    Properties:
      ContainerDefinitions:
      - Name: web
        Image: myapp:0.1
        Command:
        - bundle
        - exec
        - puma
        Environment:
        - Name: RAILS_ENV
          Value: production
        - Name: RACK_ENV
          Value: production
        PortMappings:
        - ContainerPort: 80
        LogConfiguration:
          LogDriver: awslogs
          Options:
            awslogs-region:
              Ref: AWS::Region
            awslogs-group:
              Ref: HerogateApplicationContainerLogs
            awslogs-stream-prefix: web
      - Name: worker
        Image: myapp:0.1
        Command:
        - bundle
        - exec
        - sidekiq
        Environment:
        - Name: RAILS_ENV
          Value: production
        - Name: RACK_ENV
          Value: production
        PortMappings: []
        LogConfiguration:
          LogDriver: awslogs
          Options:
            awslogs-region:
              Ref: AWS::Region
            awslogs-group:
              Ref: HerogateApplicationContainerLogs
            awslogs-stream-prefix: worker
    Type: AWS::ECS::TaskDefinition

`

	if writer.String() != expected {
		t.Fatalf("Expected template is `%s`, but get `%s`", expected, writer.String())
	}
}

func TestProcessInternalGenerateTemplate__noProcfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	app := cli.NewApp()
	writer := new(bytes.Buffer)
	app.Writer = writer

	template := `AWSTemplateFormatVersion: 2010-09-09
Description: Herogate Platform Template v1.0
Resources:
  HerogateApplicationContainer:
    Properties:
      ContainerDefinitions:
      - Image: httpd:2.4
        LogConfiguration:
          LogDriver: awslogs
          Options:
            awslogs-group:
              Ref: HerogateApplicationContainerLogs
            awslogs-region:
              Ref: AWS::Region
            awslogs-stream-prefix: web
        Name: web
        PortMappings:
        - ContainerPort: 80
    Type: AWS::ECS::TaskDefinition

`

	client := mock.NewMockClientInterface(ctrl)
	// Expect to get template
	client.EXPECT().GetTemplate("bold-art-6993").Return(template)

	processInternalGenerateTemplate(&internalGenerateTemplateContext{
		name:   "bold-art-6993",
		image:  "myapp:0.1",
		app:    app,
		client: client,
	})

	if writer.String() != template {
		t.Fatalf("Expected template is `%s`, but get `%s`", template, writer.String())
	}
}
