package container

// Definition is a CFn resource type in Herogate.
// See https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-ecs-taskdefinition-containerdefinitions.html
type Definition struct {
	Name             string            `yaml:"Name"`
	Image            string            `yaml:"Image"`
	Command          []string          `yaml:"Command"`
	Environment      []interface{}     `yaml:"Environment"` // Use `config.List()` value directly
	PortMappings     []*PortMapping    `yaml:"PortMappings"`
	LogConfiguration *LogConfiguration `yaml:"LogConfiguration"`
}

// LogConfiguration is a CFn resource type in Herogate.
// See https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-ecs-taskdefinition-containerdefinitions-logconfiguration.html
type LogConfiguration struct {
	LogDriver string                   `yaml:"LogDriver"`
	Options   *LogConfigurationOptions `yaml:"Options"`
}

// LogConfigurationOptions is awslogs driver options type.
type LogConfigurationOptions struct {
	Region       *RefObject `yaml:"awslogs-region"`
	Group        *RefObject `yaml:"awslogs-group"`
	StreamPrefix string     `yaml:"awslogs-stream-prefix"`
}

// PortMapping is a CFn resource type in Herogate.
// See https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-ecs-taskdefinition-containerdefinitions-portmappings.html
type PortMapping struct {
	ContainerPort int `yaml:"ContainerPort"`
}

// RefObject is a meta object for CFn template.
type RefObject struct {
	Ref string `yaml:"Ref"`
}

// New initializes container definition resource type for CFn by attributes.
// You can generate CFn template using `config.Set()`.
func New(name string, image string, command []string, environment []interface{}) *Definition {
	definition := &Definition{
		Name:        name,
		Image:       image,
		Command:     command,
		Environment: environment,
		LogConfiguration: &LogConfiguration{
			LogDriver: "awslogs",
			Options: &LogConfigurationOptions{
				Region:       &RefObject{Ref: "AWS::Region"},
				Group:        &RefObject{Ref: "HerogateApplicationContainerLogs"},
				StreamPrefix: name,
			},
		},
	}

	if name == "web" {
		definition.PortMappings = append(definition.PortMappings, &PortMapping{
			ContainerPort: 80,
		})
	}

	return definition
}
