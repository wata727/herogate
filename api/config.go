package api

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/sirupsen/logrus"
)

// DescribeEnvVars describes environment variables from container deifinitions.
func (c *Client) DescribeEnvVars(appName string) (map[string]string, error) {
	serviceResp, err := c.ecs.DescribeServices(&ecs.DescribeServicesInput{
		Cluster:  aws.String(appName),
		Services: []*string{aws.String(appName)},
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == ecs.ErrCodeClusterNotFoundException {
			return map[string]string{}, err
		}
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
	if len(taskResp.TaskDefinition.ContainerDefinitions) == 0 {
		return map[string]string{}, nil
	}
	definiation := taskResp.TaskDefinition.ContainerDefinitions[0]

	envVars := map[string]string{}
	for _, env := range definiation.Environment {
		envVars[aws.StringValue(env.Name)] = aws.StringValue(env.Value)
	}

	return envVars, nil
}
