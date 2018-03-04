package api

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/wata727/herogate/mock"
)

func TestDescribeEnvVars(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ecsMock := mock.NewMockECSAPI(ctrl)
	// Expected to describe services
	ecsMock.EXPECT().DescribeServices(&ecs.DescribeServicesInput{
		Cluster:  aws.String("young-eyrie-24091"),
		Services: []*string{aws.String("young-eyrie-24091")},
	}).Return(&ecs.DescribeServicesOutput{
		Services: []*ecs.Service{
			{
				TaskDefinition: aws.String("arn:aws:ecs:us-east-1:123456789:task-definition/young-eyrie-24091:1"),
				RunningCount:   aws.Int64(1),
			},
		},
	}, nil)
	// Expected to describe task definition
	ecsMock.EXPECT().DescribeTaskDefinition(&ecs.DescribeTaskDefinitionInput{
		TaskDefinition: aws.String("arn:aws:ecs:us-east-1:123456789:task-definition/young-eyrie-24091:1"),
	}).Return(&ecs.DescribeTaskDefinitionOutput{
		TaskDefinition: &ecs.TaskDefinition{
			ContainerDefinitions: []*ecs.ContainerDefinition{
				{
					Name: aws.String("web"),
					Environment: []*ecs.KeyValuePair{
						{
							Name:  aws.String("RAILS_ENV"),
							Value: aws.String("production"),
						},
						{
							Name:  aws.String("RACK_ENV"),
							Value: aws.String("production"),
						},
						{
							Name:  aws.String("SECRET_KEY_BASE"),
							Value: aws.String("011a60b8e222a55e0869e3dca9301a7736074189cb52782f1efd8b8a2e956fc44b25a6f2753f1662986c9519fbebdb7ebb4799becc75ac1a7faad0b55aee1b4b"),
						},
					},
				},
			},
		},
	}, nil)

	client := NewClient(&ClientOption{})
	client.ecs = ecsMock
	envVars, err := client.DescribeEnvVars("young-eyrie-24091")

	if err != nil {
		t.Fatal("Expected error is nil, but get error: " + err.Error())
	}

	expected := map[string]string{
		"RAILS_ENV":       "production",
		"RACK_ENV":        "production",
		"SECRET_KEY_BASE": "011a60b8e222a55e0869e3dca9301a7736074189cb52782f1efd8b8a2e956fc44b25a6f2753f1662986c9519fbebdb7ebb4799becc75ac1a7faad0b55aee1b4b",
	}

	if !cmp.Equal(expected, envVars) {
		t.Fatalf("\nDiff: %s\n", cmp.Diff(expected, envVars))
	}
}

func TestDescribeEnvVars__notFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ecsMock := mock.NewMockECSAPI(ctrl)
	// Expected to describe services
	ecsMock.EXPECT().DescribeServices(&ecs.DescribeServicesInput{
		Cluster:  aws.String("young-eyrie-24091"),
		Services: []*string{aws.String("young-eyrie-24091")},
	}).Return(nil, awserr.New(ecs.ErrCodeClusterNotFoundException, "Not found", errors.New("Not found")))

	client := NewClient(&ClientOption{})
	client.ecs = ecsMock
	_, err := client.DescribeEnvVars("young-eyrie-24091")

	if err == nil {
		t.Fatal("Expected error is not nil, but get nil as error")
	}
}

func TestSetEnvVars(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cfnMock := mock.NewMockCloudFormationAPI(ctrl)
	// Expect to call GetApp and return App
	cfnMock.EXPECT().DescribeStacks(&cloudformation.DescribeStacksInput{
		StackName: aws.String("young-eyrie-24091"),
	}).Return(&cloudformation.DescribeStacksOutput{
		Stacks: []*cloudformation.Stack{
			{
				StackStatus: aws.String("CREATE_COMPLETE"),
				Outputs: []*cloudformation.Output{
					{
						OutputKey:   aws.String("Repository"),
						OutputValue: aws.String("ssh://git-codecommit.us-east-1.amazonaws.com/v1/repos/young-eyrie-24091"),
					},
					{
						OutputKey:   aws.String("Endpoint"),
						OutputValue: aws.String("young-eyrie-24091-123456789.us-east-1.elb.amazonaws.com"),
					},
				},
				Tags: []*cloudformation.Tag{
					{
						Key:   aws.String("herogate-platform-version"),
						Value: aws.String("1.0"),
					},
				},
			},
		},
	}, nil)
	// Expect to get template with application name
	cfnMock.EXPECT().GetTemplate(&cloudformation.GetTemplateInput{
		StackName: aws.String("young-eyrie-24091"),
	}).Return(&cloudformation.GetTemplateOutput{
		TemplateBody: aws.String(`AWSTemplateFormatVersion: 2010-09-09
Description: Herogate Platform Template v1.0
Resources:
  HerogateApplicationContainer:
    Type: "AWS::ECS::TaskDefinition"
    Properties:
      ContainerDefinitions:
        - Name: web
          Image: "httpd:2.4"
`),
	}, nil)
	// Expect to update stack
	cfnMock.EXPECT().UpdateStack(&cloudformation.UpdateStackInput{
		StackName: aws.String("young-eyrie-24091"),
		TemplateBody: aws.String(`AWSTemplateFormatVersion: 2010-09-09
Description: Herogate Platform Template v1.0
Resources:
  HerogateApplicationContainer:
    Properties:
      ContainerDefinitions:
      - Environment:
        - Name: RACK_ENV
          Value: production
        - Name: RAILS_ENV
          Value: production
        Image: httpd:2.4
        Name: web
    Type: AWS::ECS::TaskDefinition
`),
		Capabilities: []*string{aws.String("CAPABILITY_NAMED_IAM")},
	})
	// Expect to wait stack update
	cfnMock.EXPECT().WaitUntilStackUpdateComplete(&cloudformation.DescribeStacksInput{
		StackName: aws.String("young-eyrie-24091"),
	}).Return(nil)

	client := NewClient(&ClientOption{})
	client.cloudFormation = cfnMock

	err := client.SetEnvVars("young-eyrie-24091", map[string]string{
		"RAILS_ENV": "production",
		"RACK_ENV":  "production",
	})
	if err != nil {
		t.Fatalf("Expected error is nil, but get `%s`", err.Error())
	}
}

func TestSetEnvVars__mergeEnvVars(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cfnMock := mock.NewMockCloudFormationAPI(ctrl)
	// Expect to call GetApp and return App
	cfnMock.EXPECT().DescribeStacks(&cloudformation.DescribeStacksInput{
		StackName: aws.String("young-eyrie-24091"),
	}).Return(&cloudformation.DescribeStacksOutput{
		Stacks: []*cloudformation.Stack{
			{
				StackStatus: aws.String("CREATE_COMPLETE"),
				Outputs: []*cloudformation.Output{
					{
						OutputKey:   aws.String("Repository"),
						OutputValue: aws.String("ssh://git-codecommit.us-east-1.amazonaws.com/v1/repos/young-eyrie-24091"),
					},
					{
						OutputKey:   aws.String("Endpoint"),
						OutputValue: aws.String("young-eyrie-24091-123456789.us-east-1.elb.amazonaws.com"),
					},
				},
				Tags: []*cloudformation.Tag{
					{
						Key:   aws.String("herogate-platform-version"),
						Value: aws.String("1.0"),
					},
				},
			},
		},
	}, nil)
	// Expect to get template with application name
	cfnMock.EXPECT().GetTemplate(&cloudformation.GetTemplateInput{
		StackName: aws.String("young-eyrie-24091"),
	}).Return(&cloudformation.GetTemplateOutput{
		TemplateBody: aws.String(`AWSTemplateFormatVersion: 2010-09-09
Description: Herogate Platform Template v1.0
Resources:
  HerogateApplicationContainer:
    Type: "AWS::ECS::TaskDefinition"
    Properties:
      ContainerDefinitions:
        - Name: web
          Image: "httpd:2.4"
          Environment:
            - Name: RACK_ENV
              Value: development
            - Name: SECRET_KEY_BASE
              Value: 011a60b8e222a55e0869e3dca9301a7736074189cb52782f1efd8b8a2e956fc44b25a6f2753f1662986c9519fbebdb7ebb4799becc75ac1a7faad0b55aee1b4b
`),
	}, nil)
	// Expect to update stack
	cfnMock.EXPECT().UpdateStack(&cloudformation.UpdateStackInput{
		StackName: aws.String("young-eyrie-24091"),
		TemplateBody: aws.String(`AWSTemplateFormatVersion: 2010-09-09
Description: Herogate Platform Template v1.0
Resources:
  HerogateApplicationContainer:
    Properties:
      ContainerDefinitions:
      - Environment:
        - Name: RACK_ENV
          Value: production
        - Name: RAILS_ENV
          Value: production
        - Name: SECRET_KEY_BASE
          Value: 011a60b8e222a55e0869e3dca9301a7736074189cb52782f1efd8b8a2e956fc44b25a6f2753f1662986c9519fbebdb7ebb4799becc75ac1a7faad0b55aee1b4b
        Image: httpd:2.4
        Name: web
    Type: AWS::ECS::TaskDefinition
`),
		Capabilities: []*string{aws.String("CAPABILITY_NAMED_IAM")},
	})
	// Expect to wait stack update
	cfnMock.EXPECT().WaitUntilStackUpdateComplete(&cloudformation.DescribeStacksInput{
		StackName: aws.String("young-eyrie-24091"),
	}).Return(nil)

	client := NewClient(&ClientOption{})
	client.cloudFormation = cfnMock

	err := client.SetEnvVars("young-eyrie-24091", map[string]string{
		"RAILS_ENV": "production",
		"RACK_ENV":  "production",
	})
	if err != nil {
		t.Fatalf("Expected error is nil, but get `%s`", err.Error())
	}
}

func TestSetEnvVars__notFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cfnMock := mock.NewMockCloudFormationAPI(ctrl)
	// Expect to call GetApp and return error
	cfnMock.EXPECT().DescribeStacks(&cloudformation.DescribeStacksInput{
		StackName: aws.String("young-eyrie-24091"),
	}).Return(nil, errors.New("Stack not found"))

	client := NewClient(&ClientOption{})
	client.cloudFormation = cfnMock

	err := client.SetEnvVars("young-eyrie-24091", map[string]string{
		"RAILS_ENV": "production",
		"RACK_ENV":  "production",
	})
	if err == nil {
		t.Fatal("Expected error is not nil, but get nil as error")
	}
}
