package api

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
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
