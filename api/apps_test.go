package api

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/golang/mock/gomock"
	"github.com/wata727/herogate/api/assets"
	"github.com/wata727/herogate/mock"
)

func TestCreateApp(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	yaml, err := assets.Asset("assets/platform.yaml")
	if err != nil {
		t.Fatal("Failed to load the template: " + err.Error())
	}

	cfnMock := mock.NewMockCloudFormationAPI(ctrl)
	// Expect to create stack with application name
	cfnMock.EXPECT().CreateStack(&cloudformation.CreateStackInput{
		StackName:    aws.String("young-eyrie-24091"),
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
	}).Return(&cloudformation.CreateStackOutput{}, nil)
	// Expect to wait stack creation
	cfnMock.EXPECT().WaitUntilStackCreateComplete(&cloudformation.DescribeStacksInput{
		StackName: aws.String("young-eyrie-24091"),
	}).Return(nil)
	// Expect to describe the stack
	cfnMock.EXPECT().DescribeStacks(&cloudformation.DescribeStacksInput{
		StackName: aws.String("young-eyrie-24091"),
	}).Return(&cloudformation.DescribeStacksOutput{
		Stacks: []*cloudformation.Stack{
			{
				StackStatus: aws.String("CREATE_COMPLETE"),
				Outputs: []*cloudformation.Output{
					{
						OutputKey:   aws.String("HerogateRepository"),
						OutputValue: aws.String("ssh://git-codecommit.us-east-1.amazonaws.com/v1/repos/young-eyrie-24091"),
					},
					{
						OutputKey:   aws.String("HerogateURL"),
						OutputValue: aws.String("young-eyrie-24091-123456789.us-east-1.elb.amazonaws.com"),
					},
				},
			},
		},
	}, nil)

	client := NewClient(&ClientOption{})
	client.cloudFormation = cfnMock

	repository, endpoint := client.CreateApp("young-eyrie-24091")
	if repository != "ssh://git-codecommit.us-east-1.amazonaws.com/v1/repos/young-eyrie-24091" {
		t.Fatalf("Expected repository is `ssh://git-codecommit.us-east-1.amazonaws.com/v1/repos/young-eyrie-24091`, but get `%s`", repository)
	}
	if endpoint != "young-eyrie-24091-123456789.us-east-1.elb.amazonaws.com" {
		t.Fatalf("Expected endpoint is `young-eyrie-24091-123456789.us-east-1.elb.amazonaws.com`, but get `%s`", endpoint)
	}
}

func TestGetAppCreationProgress(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cfnMock := mock.NewMockCloudFormationAPI(ctrl)
	cfnMock.EXPECT().ListStackResources(&cloudformation.ListStackResourcesInput{
		StackName: aws.String("young-eyrie-24091"),
	}).Return(&cloudformation.ListStackResourcesOutput{
		StackResourceSummaries: []*cloudformation.StackResourceSummary{
			{
				ResourceStatus: aws.String("CREATE_COMPLETE"),
			},
			{
				ResourceStatus: aws.String("CREATE_COMPLETE"),
			},
			{
				ResourceStatus: aws.String("CREATE_COMPLETE"),
			},
			{
				ResourceStatus: aws.String("CREATE_COMPLETE"),
			},
			{
				ResourceStatus: aws.String("CREATE_COMPLETE"),
			},
			{
				ResourceStatus: aws.String("CREATE_COMPLETE"),
			},
			{
				ResourceStatus: aws.String("CREATE_IN_PROGRESS"),
			},
			{
				ResourceStatus: aws.String("CREATE_IN_PROGRESS"),
			},
			{
				ResourceStatus: aws.String("CREATE_IN_PROGRESS"),
			},
			{
				ResourceStatus: aws.String("CREATE_IN_PROGRESS"),
			},
		},
	}, nil)

	client := NewClient(&ClientOption{})
	client.cloudFormation = cfnMock

	rate := client.GetAppCreationProgress("young-eyrie-24091")
	// Total resources: 25, Created: 6
	//   => (6 / 25) * 100 = 24
	if rate != 24 {
		t.Fatalf("Expected progress rate is `24`, but get `%d`", rate)
	}
}
