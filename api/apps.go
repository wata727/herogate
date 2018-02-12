package api

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/sirupsen/logrus"
	"github.com/wata727/herogate/api/assets"
)

//go:generate go-bindata -o assets/assets.go -pkg assets assets/platform.yaml

// CreateApp creates a new CloudFormation stack and wait until stack create complete.
// When the stack is created, returns ALB endpoint URL and CodeCommit URL.
// If the stack creation is failed, delete this stack.
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

	c.validateStackStatus(appName, stack)

	return extractStackOutput(appName, stack.Outputs)
}

func (c *Client) validateStackStatus(appName string, stack *cloudformation.Stack) {
	if aws.StringValue(stack.StackStatus) != "CREATE_COMPLETE" {
		resourcesResp, err := c.cloudFormation.ListStackResources(&cloudformation.ListStackResourcesInput{
			StackName: aws.String(appName),
		})
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"appName": appName,
			}).Fatal("Failed to get failed stack resources: " + err.Error())
		}

		// If status is not `CREATE_COMPLETE`, delete the stack.
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
}

func extractStackOutput(appName string, outputs []*cloudformation.Output) (string, string) {
	var repository, endpoint string
	for _, output := range outputs {
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
			"outputs":    outputs,
		}).Fatal("Expected outputs are not found.")
	}

	return repository, endpoint
}

// GetAppCreationProgress returns the creation progress of the application.
// This function calculates the proportion of resources that are "CREATE_COMPLETE".
func (c *Client) GetAppCreationProgress(appName string) int {
	// XXX: Count of resources of `assets/platform.yaml`
	var totalResources = 25.0

	resp, err := c.cloudFormation.ListStackResources(&cloudformation.ListStackResourcesInput{
		StackName: aws.String(appName),
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"appName": appName,
		}).Fatal("Failed to get stack resources: " + err.Error())
	}

	var created int
	for _, s := range resp.StackResourceSummaries {
		if aws.StringValue(s.ResourceStatus) == "CREATE_COMPLETE" {
			created++
		}
	}

	return int((float64(created) / totalResources) * 100)
}
