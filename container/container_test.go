package container

import (
	"testing"

	"github.com/olebedev/config"
)

func TestNew(t *testing.T) {
	definition := New(
		"worker",
		"your-app:1.0",
		"bundle exec sidekiq",
		[]interface{}{
			map[string]string{
				"Name":  "RAILS_ENV",
				"Value": "production",
			},
			map[string]string{
				"Name":  "RACK_ENV",
				"Value": "production",
			},
		},
	)

	cfg, err := config.ParseYaml("ContainerDefinitions: []")
	if err != nil {
		t.Fatal("Unexpected error occurred when generating base config: " + err.Error())
	}

	err = cfg.Set("ContainerDefinitions", []*Definition{definition})
	if err != nil {
		t.Fatal("Unexpected error occurred when setting container deifinition: " + err.Error())
	}

	yaml, err := config.RenderYaml(cfg.Root)
	if err != nil {
		t.Fatal("Unexpected error occurred when rendering YAML: " + err.Error())
	}

	expected := `ContainerDefinitions:
- Name: worker
  Image: your-app:1.0
  Command:
  - bundle exec sidekiq
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
`

	if yaml != expected {
		t.Fatalf("Expected is `%s`, but get `%s`", expected, yaml)
	}
}

func TestNew__web(t *testing.T) {
	definition := New(
		"web",
		"your-app:1.0",
		"bundle exec puma",
		[]interface{}{
			map[string]string{
				"Name":  "RAILS_ENV",
				"Value": "production",
			},
			map[string]string{
				"Name":  "RACK_ENV",
				"Value": "production",
			},
		},
	)

	cfg, err := config.ParseYaml("ContainerDefinitions: []")
	if err != nil {
		t.Fatal("Unexpected error occurred when generating base config: " + err.Error())
	}

	err = cfg.Set("ContainerDefinitions", []*Definition{definition})
	if err != nil {
		t.Fatal("Unexpected error occurred when setting container deifinition: " + err.Error())
	}

	yaml, err := config.RenderYaml(cfg.Root)
	if err != nil {
		t.Fatal("Unexpected error occurred when rendering YAML: " + err.Error())
	}

	expected := `ContainerDefinitions:
- Name: web
  Image: your-app:1.0
  Command:
  - bundle exec puma
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
`

	if yaml != expected {
		t.Fatalf("Expected is `%s`, but get `%s`", expected, yaml)
	}
}
