package api

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/sirupsen/logrus"
	"github.com/wata727/herogate/api/assets"
)

//go:generate go-bindata -o assets/assets.go -pkg assets assets/platform.yaml

func (c *Client) CreateApp(appName string) {
	yaml, err := assets.Asset("api/assets/platform.yaml")
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"appName": appName,
		}).Fatal("Failed to load the template: " + err.Error())
	}

	_, err = c.cloudFormation.CreateStack(&cloudformation.CreateStackInput{
		StackName:    aws.String(appName),
		TemplateBody: aws.String((string(yaml))),
		Parameters: []*cloudformation.Parameter{
			{
				ParameterKey:   aws.String("BuildSpec"),
				ParameterValue: aws.String(""),
			},
		},
		TimeoutInMinutes: aws.Int64(10),
		OnFailure:        aws.String("DELETE"),
		Capabilities:     []*string{aws.String("CAPABILITY_NAMED_IAM")},
		Tags: []*cloudformation.Tag{
			{
				Key:   aws.String("herogate-platform-version"),
				Value: aws.String("1.0"),
			},
		},
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"appName": appName,
		}).Fatal("Failed to create stack: " + err.Error())
	}

	// TODO: Display progress status
	c.cloudFormation.WaitUntilStackCreateComplete(&cloudformation.DescribeStacksInput{
		StackName: aws.String(appName),
	})
}
