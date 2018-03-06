package api

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/sirupsen/logrus"
	"github.com/wata727/herogate/api/assets"
	"github.com/wata727/herogate/api/objects"
)

// XXX: Count of resources of `assets/platform.yaml`
var totalResources = 27.0

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

	var repository, endpoint, platformVersion string
	for _, tag := range stack.Tags {
		if aws.StringValue(tag.Key) == "herogate-platform-version" {
			platformVersion = aws.StringValue(tag.Value)
		}
	}
	if platformVersion == "" {
		return nil, errors.New("Expected stack not found")
	}

	for _, output := range stack.Outputs {
		switch aws.StringValue(output.OutputKey) {
		case "Repository":
			repository = aws.StringValue(output.OutputValue)
		case "Endpoint":
			endpoint = "http://" + aws.StringValue(output.OutputValue) // ALB endpoint DNS doesn't contain schema
		}
	}

	return &objects.App{
		Name:            appName,
		Status:          aws.StringValue(stack.StackStatus),
		Repository:      repository,
		Endpoint:        endpoint,
		PlatformVersion: platformVersion,
	}, nil
}

// DestroyApp destroys resources in the following order:
//
// - S3 Bucket
// - ECR Repository
// - CloudFormation Stack
//
// This function waits until stack deletion complete.
func (c *Client) DestroyApp(appName string) error {
	if _, err := c.GetApp(appName); err != nil {
		return err
	}

	// At first, delete S3 bucket because if the bucket is not empty, DeleteStack is failed.
	s3Resource, err := c.cloudFormation.DescribeStackResource(&cloudformation.DescribeStackResourceInput{
		StackName:         aws.String(appName),
		LogicalResourceId: aws.String("HerogatePipelineArtifactStore"),
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"appName": appName,
		}).Debugf("Failed to get S3 bucket: " + err.Error())
	} else {
		s3Objects, err := c.s3.ListObjects(&s3.ListObjectsInput{
			Bucket: s3Resource.StackResourceDetail.PhysicalResourceId,
		})
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"appName": appName,
				"bucket":  aws.StringValue(s3Resource.StackResourceDetail.PhysicalResourceId),
			}).Debugf("Failed to list S3 objects: " + err.Error())
		} else {
			if len(s3Objects.Contents) > 0 {
				objectsToDelete := make([]*s3.ObjectIdentifier, len(s3Objects.Contents))
				for i, v := range s3Objects.Contents {
					objectsToDelete[i] = &s3.ObjectIdentifier{Key: v.Key}
				}

				_, err = c.s3.DeleteObjects(&s3.DeleteObjectsInput{
					Bucket: s3Resource.StackResourceDetail.PhysicalResourceId,
					Delete: &s3.Delete{Objects: objectsToDelete},
				})
				if err != nil {
					logrus.WithFields(logrus.Fields{
						"appName": appName,
						"bucket":  aws.StringValue(s3Resource.StackResourceDetail.PhysicalResourceId),
						"objects": objectsToDelete,
					}).Fatalf("Failed to delete S3 objects: " + err.Error())
				}
			}

			_, err = c.s3.DeleteBucket(&s3.DeleteBucketInput{
				Bucket: s3Resource.StackResourceDetail.PhysicalResourceId,
			})
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"appName": appName,
					"bucket":  aws.StringValue(s3Resource.StackResourceDetail.PhysicalResourceId),
				}).Fatalf("Failed to delete S3 bucket: " + err.Error())
			}
		}
	}

	// After that, delete ECR because if images exist, DeleteStack is failed.
	ecrResource, err := c.cloudFormation.DescribeStackResource(&cloudformation.DescribeStackResourceInput{
		StackName:         aws.String(appName),
		LogicalResourceId: aws.String("HerogateRegistry"),
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"appName": appName,
		}).Debugf("Failed to get ECR repository: " + err.Error())
	} else {
		_, err = c.ecr.DeleteRepository(&ecr.DeleteRepositoryInput{
			Force:          aws.Bool(true),
			RepositoryName: ecrResource.StackResourceDetail.PhysicalResourceId,
		})
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"appName": appName,
			}).Debugf("Failed to delete ECR repository: " + err.Error())
		}
	}

	// At last, delete CloudFormation stack.
	_, err = c.cloudFormation.DeleteStack(&cloudformation.DeleteStackInput{
		StackName: aws.String(appName),
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"appName": appName,
		}).Fatal("Failed to request for deleting stack: " + err.Error())
	}
	err = c.cloudFormation.WaitUntilStackDeleteComplete(&cloudformation.DescribeStacksInput{
		StackName: aws.String(appName),
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"appName": appName,
		}).Fatal("Failed to wait stack deletion: " + err.Error())
	}

	app, err := c.GetApp(appName)
	if err != nil {
		// Deletion success!
		return nil
	}

	// If it can get the application, check status
	if app.Status != "DELETE_COMPLETE" {
		resourcesResp, err := c.cloudFormation.ListStackResources(&cloudformation.ListStackResourcesInput{
			StackName: aws.String(app.Name),
		})
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"appName": app.Name,
			}).Fatal("Failed to get failed stack resources: " + err.Error())
		}

		logrus.WithFields(logrus.Fields{
			"appName":   app.Name,
			"summaries": resourcesResp.StackResourceSummaries,
		}).Fatal("Failed to stack creation.")
	}

	return nil
}

// GetAppDeletionProgress returns the deletion progress of the application.
// This function calculates the proportion of resources that are "DELETE_COMPLETE".
func (c *Client) GetAppDeletionProgress(appName string) int {
	resp, err := c.cloudFormation.ListStackResources(&cloudformation.ListStackResourcesInput{
		StackName: aws.String(appName),
	})
	if err != nil {
		// When the stack deleted, returns 100%
		return 100
	}

	var deleted int
	for _, s := range resp.StackResourceSummaries {
		if aws.StringValue(s.ResourceStatus) == "DELETE_COMPLETE" {
			deleted++
		}
	}

	return int((float64(deleted) / totalResources) * 100)
}

// ListApps returns applications.
func (c *Client) ListApps() []*objects.App {
	resp, err := c.cloudFormation.DescribeStacks(&cloudformation.DescribeStacksInput{})
	if err != nil {
		logrus.Fatal("Failed to describe stacks.")
	}

	apps := []*objects.App{}

	for _, stack := range resp.Stacks {
		var repository, endpoint, platformVersion string
		for _, tag := range stack.Tags {
			if aws.StringValue(tag.Key) == "herogate-platform-version" {
				platformVersion = aws.StringValue(tag.Value)
			}
		}
		if platformVersion == "" {
			continue
		}

		for _, output := range stack.Outputs {
			switch aws.StringValue(output.OutputKey) {
			case "Repository":
				repository = aws.StringValue(output.OutputValue)
			case "Endpoint":
				endpoint = "http://" + aws.StringValue(output.OutputValue) // ALB endpoint DNS doesn't contain schema
			}
		}

		apps = append(apps, &objects.App{
			Name:            aws.StringValue(stack.StackName),
			Status:          aws.StringValue(stack.StackStatus),
			Repository:      repository,
			Endpoint:        endpoint,
			PlatformVersion: platformVersion,
		})
	}

	return apps
}

// StackExists returns whether or not the stack exists.
func (c *Client) StackExists(stackName string) bool {
	resp, err := c.cloudFormation.DescribeStacks(&cloudformation.DescribeStacksInput{
		StackName: aws.String(stackName),
	})
	if err != nil {
		return false
	}
	if len(resp.Stacks) == 0 {
		return false
	}
	return true
}

// GetAppInfo returns the application info object.
// If the application not found, returns nil and error.
// The difference from `GetApp` is to include container's details, platform version, etc.
func (c *Client) GetAppInfo(appName string) (*objects.AppInfo, error) {
	app, err := c.GetApp(appName)
	if err != nil {
		return nil, err
	}

	serviceResp, err := c.ecs.DescribeServices(&ecs.DescribeServicesInput{
		Cluster:  aws.String(appName),
		Services: []*string{aws.String(appName)},
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"appName": appName,
		}).Fatal("Failed to get the ECS service: " + err.Error())
	}
	if len(serviceResp.Services) == 0 {
		logrus.WithFields(logrus.Fields{
			"appName": appName,
		}).Fatal("ECS services are not found: " + err.Error())
	}
	service := serviceResp.Services[0]

	taskResp, err := c.ecs.DescribeTaskDefinition(&ecs.DescribeTaskDefinitionInput{
		TaskDefinition: service.TaskDefinition,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"appName":  appName,
			"taskName": aws.StringValue(service.TaskDefinition),
		}).Fatal("Failed to get the ECS task definiation: " + err.Error())
	}

	containers := []*objects.Container{}
	for _, container := range taskResp.TaskDefinition.ContainerDefinitions {
		containers = append(containers, &objects.Container{
			Name:  aws.StringValue(container.Name),
			Count: aws.Int64Value(service.RunningCount),
		})
	}

	return &objects.AppInfo{
		App:        app,
		Containers: containers,
		Region:     "us-east-1", // NOTE: Currently, Fargate supported region is only `us-east-1`
	}, nil
}
