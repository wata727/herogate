package api

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/aws/aws-sdk-go/service/s3"
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
				Tags: []*cloudformation.Tag{
					{
						Key:   aws.String("herogate-platform-version"),
						Value: aws.String("1.0"),
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
	// Total resources: 27, Created: 6
	//   => (6 / 27) * 100 = 22.22...
	if rate != 22 {
		t.Fatalf("Expected progress rate is `22`, but get `%d`", rate)
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
				Tags: []*cloudformation.Tag{
					{
						Key:   aws.String("herogate-platform-version"),
						Value: aws.String("1.0"),
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

func TestGetApp_createInProgress(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cfnMock := mock.NewMockCloudFormationAPI(ctrl)
	cfnMock.EXPECT().DescribeStacks(&cloudformation.DescribeStacksInput{
		StackName: aws.String("young-eyrie-24091"),
	}).Return(&cloudformation.DescribeStacksOutput{
		Stacks: []*cloudformation.Stack{
			{
				StackStatus: aws.String("CREATE_IN_PROGRESS"),
				Tags: []*cloudformation.Tag{
					{
						Key:   aws.String("herogate-platform-version"),
						Value: aws.String("1.0"),
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
		Name:   "young-eyrie-24091",
		Status: "CREATE_IN_PROGRESS",
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

func TestDestroyApp(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cfnMock := mock.NewMockCloudFormationAPI(ctrl)
	s3Mock := mock.NewMockS3API(ctrl)
	ecrMock := mock.NewMockECRAPI(ctrl)

	// Expect to call GetApp and return App
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
				Tags: []*cloudformation.Tag{
					{
						Key:   aws.String("herogate-platform-version"),
						Value: aws.String("1.0"),
					},
				},
			},
		},
	}, nil)
	// Expect to describe S3 resource
	cfnMock.EXPECT().DescribeStackResource(&cloudformation.DescribeStackResourceInput{
		StackName:         aws.String("young-eyrie-24091"),
		LogicalResourceId: aws.String("HerogatePipelineArtifactStore"),
	}).Return(&cloudformation.DescribeStackResourceOutput{
		StackResourceDetail: &cloudformation.StackResourceDetail{
			PhysicalResourceId: aws.String("herogate-12345678-us-east-1-young-eyrie-24091"),
		},
	}, nil)
	// Expect to delete S3 bucket
	s3Mock.EXPECT().DeleteBucket(&s3.DeleteBucketInput{
		Bucket: aws.String("herogate-12345678-us-east-1-young-eyrie-24091"),
	}).Return(&s3.DeleteBucketOutput{}, nil)
	// Expect to describe ECR resource
	cfnMock.EXPECT().DescribeStackResource(&cloudformation.DescribeStackResourceInput{
		StackName:         aws.String("young-eyrie-24091"),
		LogicalResourceId: aws.String("HerogateRegistry"),
	}).Return(&cloudformation.DescribeStackResourceOutput{
		StackResourceDetail: &cloudformation.StackResourceDetail{
			PhysicalResourceId: aws.String("young-eyrie-24091"),
		},
	}, nil)
	// Expect to delete ECR repository
	ecrMock.EXPECT().DeleteRepository(&ecr.DeleteRepositoryInput{
		Force:          aws.Bool(true),
		RepositoryName: aws.String("young-eyrie-24091"),
	}).Return(&ecr.DeleteRepositoryOutput{}, nil)
	// Expect to delete CFn stack
	cfnMock.EXPECT().DeleteStack(&cloudformation.DeleteStackInput{
		StackName: aws.String("young-eyrie-24091"),
	}).Do(func(input *cloudformation.DeleteStackInput) {
		// Expect to call GetApp and return error
		cfnMock.EXPECT().DescribeStacks(&cloudformation.DescribeStacksInput{
			StackName: aws.String("young-eyrie-24091"),
		}).Return(nil, errors.New("stack not found"))
	}).Return(&cloudformation.DeleteStackOutput{}, nil)
	// Expect to wait stack deletion
	cfnMock.EXPECT().WaitUntilStackDeleteComplete(&cloudformation.DescribeStacksInput{
		StackName: aws.String("young-eyrie-24091"),
	}).Return(nil)

	client := NewClient(&ClientOption{})
	client.cloudFormation = cfnMock
	client.s3 = s3Mock
	client.ecr = ecrMock

	err := client.DestroyApp("young-eyrie-24091")
	if err != nil {
		t.Fatalf("Expected error is nil, but get `%s`", err.Error())
	}
}

func TestDestroyApp__notFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cfnMock := mock.NewMockCloudFormationAPI(ctrl)
	cfnMock.EXPECT().DescribeStacks(&cloudformation.DescribeStacksInput{
		StackName: aws.String("young-eyrie-24091"),
	}).Return(nil, errors.New("stack not found"))

	client := NewClient(&ClientOption{})
	client.cloudFormation = cfnMock

	err := client.DestroyApp("young-eyrie-24091")

	if err == nil {
		t.Fatal("Expected error is not nil, but get nil as error")
	}
}

func TestGetAppDeletionProgress(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cfnMock := mock.NewMockCloudFormationAPI(ctrl)
	cfnMock.EXPECT().ListStackResources(&cloudformation.ListStackResourcesInput{
		StackName: aws.String("young-eyrie-24091"),
	}).Return(&cloudformation.ListStackResourcesOutput{
		StackResourceSummaries: []*cloudformation.StackResourceSummary{
			{
				ResourceStatus: aws.String("DELETE_COMPLETE"),
			},
			{
				ResourceStatus: aws.String("DELETE_COMPLETE"),
			},
			{
				ResourceStatus: aws.String("DELETE_COMPLETE"),
			},
			{
				ResourceStatus: aws.String("DELETE_COMPLETE"),
			},
			{
				ResourceStatus: aws.String("DELETE_IN_PROGRESS"),
			},
			{
				ResourceStatus: aws.String("DELETE_IN_PROGRESS"),
			},
			{
				ResourceStatus: aws.String("DELETE_IN_PROGRESS"),
			},
			{
				ResourceStatus: aws.String("DELETE_IN_PROGRESS"),
			},
			{
				ResourceStatus: aws.String("DELETE_IN_PROGRESS"),
			},
			{
				ResourceStatus: aws.String("DELETE_IN_PROGRESS"),
			},
		},
	}, nil)

	client := NewClient(&ClientOption{})
	client.cloudFormation = cfnMock

	rate := client.GetAppDeletionProgress("young-eyrie-24091")
	// Total resources: 27, deleted: 4
	//   => (4 / 27) * 100 = 14.81...
	if rate != 14 {
		t.Fatalf("Expected progress rate is `14`, but get `%d`", rate)
	}
}

func TestGetAppDeletionProgress__notFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cfnMock := mock.NewMockCloudFormationAPI(ctrl)
	cfnMock.EXPECT().ListStackResources(&cloudformation.ListStackResourcesInput{
		StackName: aws.String("young-eyrie-24091"),
	}).Return(nil, errors.New("stack not found"))

	client := NewClient(&ClientOption{})
	client.cloudFormation = cfnMock

	rate := client.GetAppDeletionProgress("young-eyrie-24091")
	if rate != 100 {
		t.Fatalf("Expected progress rate is `100`, but get `%d`", rate)
	}
}

func TestListApps(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cfnMock := mock.NewMockCloudFormationAPI(ctrl)
	cfnMock.EXPECT().DescribeStacks(&cloudformation.DescribeStacksInput{}).Return(&cloudformation.DescribeStacksOutput{
		Stacks: []*cloudformation.Stack{
			{
				StackName:   aws.String("young-eyrie-24091"),
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
				Tags: []*cloudformation.Tag{
					{
						Key:   aws.String("herogate-platform-version"),
						Value: aws.String("1.0"),
					},
				},
			},
			{
				StackName:   aws.String("proud-lab-1661"),
				StackStatus: aws.String("CREATE_IN_PROGRESS"),
				Tags: []*cloudformation.Tag{
					{
						Key:   aws.String("herogate-platform-version"),
						Value: aws.String("1.0"),
					},
				},
			},
		},
	}, nil)

	client := NewClient(&ClientOption{})
	client.cloudFormation = cfnMock
	apps := client.ListApps()

	expected := []*objects.App{
		{
			Name:       "young-eyrie-24091",
			Status:     "CREATE_COMPLETE",
			Repository: "ssh://git-codecommit.us-east-1.amazonaws.com/v1/repos/young-eyrie-24091",
			Endpoint:   "http://young-eyrie-24091-123456789.us-east-1.elb.amazonaws.com",
		},
		{
			Name:   "proud-lab-1661",
			Status: "CREATE_IN_PROGRESS",
		},
	}

	if !cmp.Equal(expected, apps) {
		t.Fatalf("\nDiff: %s\n", cmp.Diff(expected, apps))
	}
}

func TestStackExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cfnMock := mock.NewMockCloudFormationAPI(ctrl)
	cfnMock.EXPECT().DescribeStacks(&cloudformation.DescribeStacksInput{
		StackName: aws.String("young-eyrie-24091"),
	}).Return(&cloudformation.DescribeStacksOutput{
		Stacks: []*cloudformation.Stack{
			{
				StackStatus: aws.String("CREATE_COMPLETE"),
			},
		},
	}, nil)

	client := NewClient(&ClientOption{})
	client.cloudFormation = cfnMock

	if !client.StackExists("young-eyrie-24091") {
		t.Fatal("Expected to exists the stack, but did not exist.")
	}
}

func TestStackExists__notFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cfnMock := mock.NewMockCloudFormationAPI(ctrl)
	cfnMock.EXPECT().DescribeStacks(&cloudformation.DescribeStacksInput{
		StackName: aws.String("young-eyrie-24091"),
	}).Return(&cloudformation.DescribeStacksOutput{
		Stacks: []*cloudformation.Stack{},
	}, nil)

	client := NewClient(&ClientOption{})
	client.cloudFormation = cfnMock

	if client.StackExists("young-eyrie-24091") {
		t.Fatal("Expected to not exists the stack, but exists.")
	}
}
