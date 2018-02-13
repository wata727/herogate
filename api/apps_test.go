package api

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/wata727/herogate/api/assets"
	"github.com/wata727/herogate/api/objects"
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
		StackName:        aws.String("young-eyrie-24091"),
		TemplateBody:     aws.String((string(yaml))),
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
						OutputKey:   aws.String("Repository"),
						OutputValue: aws.String("ssh://git-codecommit.us-east-1.amazonaws.com/v1/repos/young-eyrie-24091"),
					},
					{
						OutputKey:   aws.String("Endpoint"),
						OutputValue: aws.String("young-eyrie-24091-123456789.us-east-1.elb.amazonaws.com"),
					},
				},
			},
		},
	}, nil)

	client := NewClient(&ClientOption{})
	client.cloudFormation = cfnMock

	app := client.CreateApp("young-eyrie-24091")
	expected := &objects.App{
		Name:       "young-eyrie-24091",
		Status:     "CREATE_COMPLETE",
		Repository: "ssh://git-codecommit.us-east-1.amazonaws.com/v1/repos/young-eyrie-24091",
		Endpoint:   "http://young-eyrie-24091-123456789.us-east-1.elb.amazonaws.com",
	}
	if !cmp.Equal(expected, app) {
		t.Fatalf("\nDiff: %s\n", cmp.Diff(expected, app))
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
	// Total resources: 26, Created: 6
	//   => (6 / 26) * 100 = 23.07...
	if rate != 23 {
		t.Fatalf("Expected progress rate is `23`, but get `%d`", rate)
	}
}

func TestGetApp(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cfnMock := mock.NewMockCloudFormationAPI(ctrl)
	cfnMock.EXPECT().DescribeStacks(&cloudformation.DescribeStacksInput{
		StackName: aws.String("young-eyrie-24091"),
	}).Return(&cloudformation.DescribeStacksOutput{
		Stacks: []*cloudformation.Stack{
			{
				StackStatus: aws.String("CREATE_COMPLETE"),
				Outputs: []*cloudformation.Output{
					{
						OutputKey:   aws.String("Repository"),
						OutputValue: aws.String("ssh://git-codecommit.us-east-1.amazonaws.com/v1/repos/young-eyrie-24091"),
					},
					{
						OutputKey:   aws.String("Endpoint"),
						OutputValue: aws.String("young-eyrie-24091-123456789.us-east-1.elb.amazonaws.com"),
					},
				},
			},
		},
	}, nil)

	client := NewClient(&ClientOption{})
	client.cloudFormation = cfnMock
	app, err := client.GetApp("young-eyrie-24091")

	if err != nil {
		t.Fatal("Expected error is nil, but get error: " + err.Error())
	}

	expected := &objects.App{
		Name:       "young-eyrie-24091",
		Status:     "CREATE_COMPLETE",
		Repository: "ssh://git-codecommit.us-east-1.amazonaws.com/v1/repos/young-eyrie-24091",
		Endpoint:   "http://young-eyrie-24091-123456789.us-east-1.elb.amazonaws.com",
	}
	if !cmp.Equal(expected, app) {
		t.Fatalf("\nDiff: %s\n", cmp.Diff(expected, app))
	}
}

func TestGetApp__notFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cfnMock := mock.NewMockCloudFormationAPI(ctrl)
	cfnMock.EXPECT().DescribeStacks(&cloudformation.DescribeStacksInput{
		StackName: aws.String("young-eyrie-24091"),
	}).Return(nil, errors.New("stack not found"))

	client := NewClient(&ClientOption{})
	client.cloudFormation = cfnMock
	app, err := client.GetApp("young-eyrie-24091")

	if err == nil {
		t.Fatal("Expected error is not nil, but get nil as error")
	}
	if app != nil {
		t.Fatal("Expected app is nil, but get app")
	}
}
