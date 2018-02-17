package api

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/sirupsen/logrus"
	"github.com/wata727/herogate/api/assets"
	"github.com/wata727/herogate/api/objects"
)

//go:generate go-bindata -o assets/assets.go -pkg assets assets/platform.yaml

// CreateApp creates a new CloudFormation stack and wait until stack create complete.
// When the stack is created, returns ALB endpoint URL and CodeCommit URL.
// If the stack creation is failed, delete this stack.
func (c *Client) CreateApp(appName string) *objects.App {
	yaml, err := assets.Asset("assets/platform.yaml")
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"appName": appName,
		}).Fatal("Failed to load the template: " + err.Error())
	}

	_, err = c.cloudFormation.CreateStack(&cloudformation.CreateStackInput{
		StackName:        aws.String(appName),
		TemplateBody:     aws.String((string(yaml))),
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

	app, err := c.GetApp(appName)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"appName": appName,
		}).Fatal("Failed to get app: " + err.Error())
	}

	c.validateAppStatus(app)

	return app
}

func (c *Client) validateAppStatus(app *objects.App) {
	if app.Status != "CREATE_COMPLETE" {
		resourcesResp, err := c.cloudFormation.ListStackResources(&cloudformation.ListStackResourcesInput{
			StackName: aws.String(app.Name),
		})
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"appName": app.Name,
			}).Fatal("Failed to get failed stack resources: " + err.Error())
		}

		// If status is not `CREATE_COMPLETE`, delete the stack.
		_, err = c.cloudFormation.DeleteStack(&cloudformation.DeleteStackInput{
			StackName: aws.String(app.Name),
		})
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"appName": app.Name,
			}).Fatal("Failed to request for deleting stack: " + err.Error())
		}

		logrus.WithFields(logrus.Fields{
			"appName":   app.Name,
			"summaries": resourcesResp.StackResourceSummaries,
		}).Fatal("Failed to stack creation.")
	}
}

// GetAppCreationProgress returns the creation progress of the application.
// This function calculates the proportion of resources that are "CREATE_COMPLETE".
func (c *Client) GetAppCreationProgress(appName string) int {
	// XXX: Count of resources of `assets/platform.yaml`
	var totalResources = 27.0

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

// GetApp returns the application object.
// If the application not found, returns nil and error.
func (c *Client) GetApp(appName string) (*objects.App, error) {
	resp, err := c.cloudFormation.DescribeStacks(&cloudformation.DescribeStacksInput{
		StackName: aws.String(appName),
	})
	if err != nil {
		return nil, err
	}
	if len(resp.Stacks) == 0 {
		return nil, errors.New("Expected stack not found")
	}
	stack := resp.Stacks[0]

	var repository, endpoint string
	for _, output := range stack.Outputs {
		switch aws.StringValue(output.OutputKey) {
		case "Repository":
			repository = aws.StringValue(output.OutputValue)
		case "Endpoint":
			endpoint = aws.StringValue(output.OutputValue)
		}
	}

	if aws.StringValue(stack.StackStatus) == "CREATE_COMPLETE" {
		if repository == "" || endpoint == "" {
			logrus.WithFields(logrus.Fields{
				"appName":    appName,
				"repository": repository,
				"endpoint":   endpoint,
				"outputs":    stack.Outputs,
			}).Fatal("Expected outputs are not found.")
		}
	}

	return &objects.App{
		Name:       appName,
		Status:     aws.StringValue(stack.StackStatus),
		Repository: repository,
		Endpoint:   "http://" + endpoint, // ALB endpoint DNS doesn't contain schema
	}, nil
}
