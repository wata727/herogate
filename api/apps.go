package api

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/sirupsen/logrus"
	"github.com/wata727/herogate/api/assets"
)

//go:generate go-bindata -o assets/assets.go -pkg assets assets/platform.yaml

func (c *Client) CreateApp(appName string) (string, string) {
	yaml, err := assets.Asset("assets/platform.yaml")
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
		}).Fatal("Failed to request for creating stack: " + err.Error())
	}

	err = c.cloudFormation.WaitUntilStackCreateComplete(&cloudformation.DescribeStacksInput{
		StackName: aws.String(appName),
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"appName": appName,
		}).Fatal("Failed to wait stack creation: " + err.Error())
	}

	resp, err := c.cloudFormation.DescribeStacks(&cloudformation.DescribeStacksInput{
		StackName: aws.String(appName),
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"appName": appName,
		}).Fatal("Failed to get created stack information: " + err.Error())
	}
	if len(resp.Stacks) == 0 {
		logrus.WithFields(logrus.Fields{
			"appName": appName,
		}).Fatal("Expected stack not found.")
	}
	stack := resp.Stacks[0]

	if aws.StringValue(stack.StackStatus) != "CREATE_COMPLETE" {
		resourcesResp, err := c.cloudFormation.ListStackResources(&cloudformation.ListStackResourcesInput{
			StackName: aws.String(appName),
		})
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"appName": appName,
			}).Fatal("Failed to get failed stack resources: " + err.Error())
		}

		_, err = c.cloudFormation.DeleteStack(&cloudformation.DeleteStackInput{
			StackName: aws.String(appName),
		})
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"appName": appName,
			}).Fatal("Failed to request for deleting stack: " + err.Error())
		}

		logrus.WithFields(logrus.Fields{
			"appName":   appName,
			"summaries": resourcesResp.StackResourceSummaries,
		}).Fatal("Failed to stack creation.")
	}

	var repository, endpoint string
	for _, output := range stack.Outputs {
		switch aws.StringValue(output.OutputKey) {
		case "HerogateRepository":
			repository = aws.StringValue(output.OutputValue)
		case "HerogateURL":
			endpoint = aws.StringValue(output.OutputValue)
		}
	}

	if repository == "" || endpoint == "" {
		logrus.WithFields(logrus.Fields{
			"appName":    appName,
			"repository": repository,
			"endpoint":   endpoint,
			"outputs":    stack.Outputs,
		}).Fatal("Expected outputs are not found.")
	}

	return repository, endpoint
}

func (c *Client) GetAppCreationProgress(appName string) int {
	// XXX: Count of resources of `assets/platform.yaml`
	var totalResources int = 25

	resp, err := c.cloudFormation.ListStackResources(&cloudformation.ListStackResourcesInput{
		StackName: aws.String(appName),
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"appName": appName,
		}).Fatal("Failed to get stack resources: " + err.Error())
	}

	var created int = 0
	for _, s := range resp.StackResourceSummaries {
		if aws.StringValue(s.ResourceStatus) == "CREATE_COMPLETE" {
			created += 1
		}
	}

	return int((float64(created) / float64(totalResources)) * 100)
}
