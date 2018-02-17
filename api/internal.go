package api

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/sirupsen/logrus"
)

// GetTemplate is wrapper for cloudformation.GetTemplate
func (c *Client) GetTemplate(appName string) string {
	resp, err := c.cloudFormation.GetTemplate(&cloudformation.GetTemplateInput{
		StackName: aws.String(appName),
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"appName": appName,
		}).Fatal("Failed to get stack template: " + err.Error())
	}

	return aws.StringValue(resp.TemplateBody)
}
