package api

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/golang/mock/gomock"
	"github.com/wata727/herogate/mock"
)

func TestGetTemplate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cfnMock := mock.NewMockCloudFormationAPI(ctrl)
	// Expect to get template with application name
	cfnMock.EXPECT().GetTemplate(&cloudformation.GetTemplateInput{
		StackName: aws.String("bold-art-6993"),
	}).Return(&cloudformation.GetTemplateOutput{
		TemplateBody: aws.String(`
AWSTemplateFormatVersion: 2010-09-09
Description: Herogate Platform Template v1.0
`),
	}, nil)

	client := NewClient(&ClientOption{})
	client.cloudFormation = cfnMock

	template := client.GetTemplate("bold-art-6993")
	expected := `
AWSTemplateFormatVersion: 2010-09-09
Description: Herogate Platform Template v1.0
`

	if template != expected {
		t.Fatalf("Expected template is `%s`, but get `%s`", expected, template)
	}
}
