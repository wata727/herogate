package api

import (
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/wata727/herogate/api/options"
	"github.com/wata727/herogate/log"
	"github.com/wata727/herogate/mock"
)

func TestDescribeLogs__noConfig(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := NewClient(&ClientOption{})
	client.codeBuild = mockCodeBuild(ctrl)
	client.cloudWatchLogs = mockCloudWatchLogs(ctrl)
	client.ecs = mockECS(ctrl)

	expected := []*log.Log{
		{
			ID:        "a990c8e1-7190-463f-af65-49446c23741c",
			Timestamp: time.Date(2018, time.February, 2, 11, 0, 5, 0, time.FixedZone("UTC", 0)),
			Source:    "herogate",
			Process:   "deployer",
			Message:   "(service TestApp) has reached a steady state.",
		},
		{
			ID:        "TestApp:d6940abd-ba2c-4e36-b124-1c3d81f9ee26-1517621401000-[Container] 2018/01/26 18:20:01 Waiting for agent ping\n",
			Timestamp: time.Date(2018, time.February, 3, 1, 30, 1, 0, time.FixedZone("UTC", 0)),
			Source:    "herogate",
			Process:   "builder",
			Message:   "[Container] 2018/01/26 18:20:01 Waiting for agent ping",
		},
		{
			ID:        "TestApp:d6940abd-ba2c-4e36-b124-1c3d81f9ee26-1517621482000-[Container] 2018/01/26 18:20:04 Phase context status code:  Message: \n",
			Timestamp: time.Date(2018, time.February, 3, 1, 31, 22, 0, time.FixedZone("UTC", 0)),
			Source:    "herogate",
			Process:   "builder",
			Message:   "[Container] 2018/01/26 18:20:04 Phase context status code:  Message: ",
		},
		{
			ID:        "8720a9e8-2a5a-4f83-8b01-d9fc740fa6e4",
			Timestamp: time.Date(2018, time.February, 3, 1, 32, 22, 0, time.FixedZone("UTC", 0)),
			Source:    "herogate",
			Process:   "deployer",
			Message:   "(service TestApp) has started 1 running tasks: (task 2cf5252f-4b9e-48c3-ba73-76c1aa42e323)",
		},
		{
			ID:        "5bd5b863-72e8-4f51-a255-33c7c0721345",
			Timestamp: time.Date(2018, time.February, 3, 1, 34, 56, 0, time.FixedZone("UTC", 0)),
			Source:    "herogate",
			Process:   "deployer",
			Message:   "(service TestApp) has stopped 1 running tasks: (task 2cf5252f-4b9e-48c3-ba73-76c1aa42e323)",
		},
		{
			ID:        "354a98fa-8c77-4dc6-9c43-1ca33f293ea4",
			Timestamp: time.Date(2018, time.February, 3, 1, 35, 10, 0, time.FixedZone("UTC", 0)),
			Source:    "herogate",
			Process:   "deployer",
			Message:   "(service TestApp) has reached a steady state.",
		},
	}

	logs, err := client.DescribeLogs("TestApp", &options.DescribeLogs{})
	if err != nil {
		t.Fatalf("Expected error is nil, but get `%s`", err.Error())
	}
	if !cmp.Equal(expected, logs) {
		t.Fatalf("\nDiff: %s\n", cmp.Diff(expected, logs))
	}
}

func TestDescribeLogs__sourceHerogate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := NewClient(&ClientOption{})
	client.codeBuild = mockCodeBuild(ctrl)
	client.cloudWatchLogs = mockCloudWatchLogs(ctrl)
	client.ecs = mockECS(ctrl)

	expected := []*log.Log{
		{
			ID:        "a990c8e1-7190-463f-af65-49446c23741c",
			Timestamp: time.Date(2018, time.February, 2, 11, 0, 5, 0, time.FixedZone("UTC", 0)),
			Source:    "herogate",
			Process:   "deployer",
			Message:   "(service TestApp) has reached a steady state.",
		},
		{
			ID:        "TestApp:d6940abd-ba2c-4e36-b124-1c3d81f9ee26-1517621401000-[Container] 2018/01/26 18:20:01 Waiting for agent ping\n",
			Timestamp: time.Date(2018, time.February, 3, 1, 30, 1, 0, time.FixedZone("UTC", 0)),
			Source:    "herogate",
			Process:   "builder",
			Message:   "[Container] 2018/01/26 18:20:01 Waiting for agent ping",
		},
		{
			ID:        "TestApp:d6940abd-ba2c-4e36-b124-1c3d81f9ee26-1517621482000-[Container] 2018/01/26 18:20:04 Phase context status code:  Message: \n",
			Timestamp: time.Date(2018, time.February, 3, 1, 31, 22, 0, time.FixedZone("UTC", 0)),
			Source:    "herogate",
			Process:   "builder",
			Message:   "[Container] 2018/01/26 18:20:04 Phase context status code:  Message: ",
		},
		{
			ID:        "8720a9e8-2a5a-4f83-8b01-d9fc740fa6e4",
			Timestamp: time.Date(2018, time.February, 3, 1, 32, 22, 0, time.FixedZone("UTC", 0)),
			Source:    "herogate",
			Process:   "deployer",
			Message:   "(service TestApp) has started 1 running tasks: (task 2cf5252f-4b9e-48c3-ba73-76c1aa42e323)",
		},
		{
			ID:        "5bd5b863-72e8-4f51-a255-33c7c0721345",
			Timestamp: time.Date(2018, time.February, 3, 1, 34, 56, 0, time.FixedZone("UTC", 0)),
			Source:    "herogate",
			Process:   "deployer",
			Message:   "(service TestApp) has stopped 1 running tasks: (task 2cf5252f-4b9e-48c3-ba73-76c1aa42e323)",
		},
		{
			ID:        "354a98fa-8c77-4dc6-9c43-1ca33f293ea4",
			Timestamp: time.Date(2018, time.February, 3, 1, 35, 10, 0, time.FixedZone("UTC", 0)),
			Source:    "herogate",
			Process:   "deployer",
			Message:   "(service TestApp) has reached a steady state.",
		},
	}

	logs, err := client.DescribeLogs("TestApp", &options.DescribeLogs{Source: "herogate"})
	if err != nil {
		t.Fatalf("Expected error is nil, but get `%s`", err.Error())
	}
	if !cmp.Equal(expected, logs) {
		t.Fatalf("\nDiff: %s\n", cmp.Diff(expected, logs))
	}
}

func TestDescribeLogs__processBuilder(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := NewClient(&ClientOption{})
	client.codeBuild = mockCodeBuild(ctrl)
	client.cloudWatchLogs = mockCloudWatchLogs(ctrl)

	expected := []*log.Log{
		{
			ID:        "TestApp:d6940abd-ba2c-4e36-b124-1c3d81f9ee26-1517621401000-[Container] 2018/01/26 18:20:01 Waiting for agent ping\n",
			Timestamp: time.Date(2018, time.February, 3, 1, 30, 1, 0, time.FixedZone("UTC", 0)),
			Source:    "herogate",
			Process:   "builder",
			Message:   "[Container] 2018/01/26 18:20:01 Waiting for agent ping",
		},
		{
			ID:        "TestApp:d6940abd-ba2c-4e36-b124-1c3d81f9ee26-1517621482000-[Container] 2018/01/26 18:20:04 Phase context status code:  Message: \n",
			Timestamp: time.Date(2018, time.February, 3, 1, 31, 22, 0, time.FixedZone("UTC", 0)),
			Source:    "herogate",
			Process:   "builder",
			Message:   "[Container] 2018/01/26 18:20:04 Phase context status code:  Message: ",
		},
	}

	logs, err := client.DescribeLogs("TestApp", &options.DescribeLogs{Process: "builder"})
	if err != nil {
		t.Fatalf("Expected error is nil, but get `%s`", err.Error())
	}
	if !cmp.Equal(expected, logs) {
		t.Fatalf("\nDiff: %s\n", cmp.Diff(expected, logs))
	}
}

func TestDescribeLogs__processDeployer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := NewClient(&ClientOption{})
	client.ecs = mockECS(ctrl)

	expected := []*log.Log{
		{
			ID:        "a990c8e1-7190-463f-af65-49446c23741c",
			Timestamp: time.Date(2018, time.February, 2, 11, 0, 5, 0, time.FixedZone("UTC", 0)),
			Source:    "herogate",
			Process:   "deployer",
			Message:   "(service TestApp) has reached a steady state.",
		},
		{
			ID:        "8720a9e8-2a5a-4f83-8b01-d9fc740fa6e4",
			Timestamp: time.Date(2018, time.February, 3, 1, 32, 22, 0, time.FixedZone("UTC", 0)),
			Source:    "herogate",
			Process:   "deployer",
			Message:   "(service TestApp) has started 1 running tasks: (task 2cf5252f-4b9e-48c3-ba73-76c1aa42e323)",
		},
		{
			ID:        "5bd5b863-72e8-4f51-a255-33c7c0721345",
			Timestamp: time.Date(2018, time.February, 3, 1, 34, 56, 0, time.FixedZone("UTC", 0)),
			Source:    "herogate",
			Process:   "deployer",
			Message:   "(service TestApp) has stopped 1 running tasks: (task 2cf5252f-4b9e-48c3-ba73-76c1aa42e323)",
		},
		{
			ID:        "354a98fa-8c77-4dc6-9c43-1ca33f293ea4",
			Timestamp: time.Date(2018, time.February, 3, 1, 35, 10, 0, time.FixedZone("UTC", 0)),
			Source:    "herogate",
			Process:   "deployer",
			Message:   "(service TestApp) has reached a steady state.",
		},
	}

	logs, err := client.DescribeLogs("TestApp", &options.DescribeLogs{Process: "deployer"})
	if err != nil {
		t.Fatalf("Expected error is nil, but get `%s`", err.Error())
	}
	if !cmp.Equal(expected, logs) {
		t.Fatalf("\nDiff: %s\n", cmp.Diff(expected, logs))
	}
}

func TestDescribeLogs__sourceHerogate__processBuilder(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := NewClient(&ClientOption{})
	client.codeBuild = mockCodeBuild(ctrl)
	client.cloudWatchLogs = mockCloudWatchLogs(ctrl)

	expected := []*log.Log{
		{
			ID:        "TestApp:d6940abd-ba2c-4e36-b124-1c3d81f9ee26-1517621401000-[Container] 2018/01/26 18:20:01 Waiting for agent ping\n",
			Timestamp: time.Date(2018, time.February, 3, 1, 30, 1, 0, time.FixedZone("UTC", 0)),
			Source:    "herogate",
			Process:   "builder",
			Message:   "[Container] 2018/01/26 18:20:01 Waiting for agent ping",
		},
		{
			ID:        "TestApp:d6940abd-ba2c-4e36-b124-1c3d81f9ee26-1517621482000-[Container] 2018/01/26 18:20:04 Phase context status code:  Message: \n",
			Timestamp: time.Date(2018, time.February, 3, 1, 31, 22, 0, time.FixedZone("UTC", 0)),
			Source:    "herogate",
			Process:   "builder",
			Message:   "[Container] 2018/01/26 18:20:04 Phase context status code:  Message: ",
		},
	}

	logs, err := client.DescribeLogs("TestApp", &options.DescribeLogs{
		Source:  "herogate",
		Process: "builder",
	})
	if err != nil {
		t.Fatalf("Expected error is nil, but get `%s`", err.Error())
	}
	if !cmp.Equal(expected, logs) {
		t.Fatalf("\nDiff: %s\n", cmp.Diff(expected, logs))
	}
}

func TestDescribeLogs__sourceHerogate__processDeployer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := NewClient(&ClientOption{})
	client.ecs = mockECS(ctrl)

	expected := []*log.Log{
		{
			ID:        "a990c8e1-7190-463f-af65-49446c23741c",
			Timestamp: time.Date(2018, time.February, 2, 11, 0, 5, 0, time.FixedZone("UTC", 0)),
			Source:    "herogate",
			Process:   "deployer",
			Message:   "(service TestApp) has reached a steady state.",
		},
		{
			ID:        "8720a9e8-2a5a-4f83-8b01-d9fc740fa6e4",
			Timestamp: time.Date(2018, time.February, 3, 1, 32, 22, 0, time.FixedZone("UTC", 0)),
			Source:    "herogate",
			Process:   "deployer",
			Message:   "(service TestApp) has started 1 running tasks: (task 2cf5252f-4b9e-48c3-ba73-76c1aa42e323)",
		},
		{
			ID:        "5bd5b863-72e8-4f51-a255-33c7c0721345",
			Timestamp: time.Date(2018, time.February, 3, 1, 34, 56, 0, time.FixedZone("UTC", 0)),
			Source:    "herogate",
			Process:   "deployer",
			Message:   "(service TestApp) has stopped 1 running tasks: (task 2cf5252f-4b9e-48c3-ba73-76c1aa42e323)",
		},
		{
			ID:        "354a98fa-8c77-4dc6-9c43-1ca33f293ea4",
			Timestamp: time.Date(2018, time.February, 3, 1, 35, 10, 0, time.FixedZone("UTC", 0)),
			Source:    "herogate",
			Process:   "deployer",
			Message:   "(service TestApp) has reached a steady state.",
		},
	}

	logs, err := client.DescribeLogs("TestApp", &options.DescribeLogs{
		Source:  "herogate",
		Process: "deployer",
	})
	if err != nil {
		t.Fatalf("Expected error is nil, but get `%s`", err.Error())
	}
	if !cmp.Equal(expected, logs) {
		t.Fatalf("\nDiff: %s\n", cmp.Diff(expected, logs))
	}
}

func TestDescribeLogs__sourceInvalid(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := NewClient(&ClientOption{})

	expected := []*log.Log{}

	logs, err := client.DescribeLogs("TestApp", &options.DescribeLogs{Source: "invalid"})
	if err != nil {
		t.Fatalf("Expected error is nil, but get `%s`", err.Error())
	}
	if !cmp.Equal(expected, logs) {
		t.Fatalf("\nDiff: %s\n", cmp.Diff(expected, logs))
	}
}

func TestDescribeLogs__processInvalid(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := NewClient(&ClientOption{})

	expected := []*log.Log{}

	logs, err := client.DescribeLogs("TestApp", &options.DescribeLogs{Process: "invalid"})
	if err != nil {
		t.Fatalf("Expected error is nil, but get `%s`", err.Error())
	}
	if !cmp.Equal(expected, logs) {
		t.Fatalf("\nDiff: %s\n", cmp.Diff(expected, logs))
	}
}

func TestDescribeLogs__sourceHerogate__processInvalid(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := NewClient(&ClientOption{})

	expected := []*log.Log{}

	logs, err := client.DescribeLogs("TestApp", &options.DescribeLogs{
		Source:  "herogate",
		Process: "invalid",
	})
	if err != nil {
		t.Fatalf("Expected error is nil, but get `%s`", err.Error())
	}
	if !cmp.Equal(expected, logs) {
		t.Fatalf("\nDiff: %s\n", cmp.Diff(expected, logs))
	}
}

func TestDescribeLogs__projectNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := NewClient(&ClientOption{})
	client.codeBuild = mockCodeBuildNotFound(ctrl)

	expected := []*log.Log{}

	logs, err := client.DescribeLogs("TestApp", &options.DescribeLogs{})
	if err == nil {
		t.Fatalf("Expected error is ErrCodeResourceNotFoundException, but get `%s`", err.Error())
	}
	if !cmp.Equal(expected, logs) {
		t.Fatalf("\nDiff: %s\n", cmp.Diff(expected, logs))
	}
}

func TestDescribeLogs__serviceNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := NewClient(&ClientOption{})
	client.codeBuild = mockCodeBuild(ctrl)
	client.cloudWatchLogs = mockCloudWatchLogs(ctrl)
	client.ecs = mockECSNotFound(ctrl)

	expected := []*log.Log{}

	logs, err := client.DescribeLogs("TestApp", &options.DescribeLogs{})
	if err == nil {
		t.Fatalf("Expected error is ErrCodeClusterNotFoundException, but get `%s`", err.Error())
	}
	if !cmp.Equal(expected, logs) {
		t.Fatalf("\nDiff: %s\n", cmp.Diff(expected, logs))
	}
}

// Mock functions
func mockCodeBuild(ctrl *gomock.Controller) *mock.MockCodeBuildAPI {
	codeBuildMock := mock.NewMockCodeBuildAPI(ctrl)

	// Mock codebuild.ListBuildsForProject
	codeBuildMock.EXPECT().ListBuildsForProject(&codebuild.ListBuildsForProjectInput{
		ProjectName: aws.String("TestApp"),
	}).Return(&codebuild.ListBuildsForProjectOutput{
		Ids: []*string{
			aws.String("TestApp:d6940abd-ba2c-4e36-b124-1c3d81f9ee26"),
			aws.String("TestApp:b3a92742-28f2-4c11-a5bc-495311631d6d"),
			aws.String("TestApp:eceb5888-72e4-4f02-b8a3-2ecfa37bf785"),
		},
	}, nil)

	// Mock codebuild.BatchGetBuilds
	codeBuildMock.EXPECT().BatchGetBuilds(&codebuild.BatchGetBuildsInput{
		Ids: []*string{aws.String("TestApp:d6940abd-ba2c-4e36-b124-1c3d81f9ee26")},
	}).Return(&codebuild.BatchGetBuildsOutput{
		Builds: []*codebuild.Build{
			{
				Logs: &codebuild.LogsLocation{
					GroupName:  aws.String("/aws/codebuild/TestApp"),
					StreamName: aws.String("d6940abd-ba2c-4e36-b124-1c3d81f9ee26"),
				},
			},
		},
	}, nil)

	return codeBuildMock
}

func mockCodeBuildNotFound(ctrl *gomock.Controller) *mock.MockCodeBuildAPI {
	codeBuildMock := mock.NewMockCodeBuildAPI(ctrl)

	codeBuildMock.EXPECT().ListBuildsForProject(&codebuild.ListBuildsForProjectInput{
		ProjectName: aws.String("TestApp"),
	}).Return(nil, awserr.New(codebuild.ErrCodeResourceNotFoundException, "Not found", errors.New("Not found")))

	return codeBuildMock
}

func mockCloudWatchLogs(ctrl *gomock.Controller) *mock.MockCloudWatchLogsAPI {
	cloudWatchLogsMock := mock.NewMockCloudWatchLogsAPI(ctrl)

	// Mock cloudwatchlogs.GetLogEvents
	cloudWatchLogsMock.EXPECT().GetLogEvents(&cloudwatchlogs.GetLogEventsInput{
		LogGroupName:  aws.String("/aws/codebuild/TestApp"),
		LogStreamName: aws.String("d6940abd-ba2c-4e36-b124-1c3d81f9ee26"),
	}).Return(&cloudwatchlogs.GetLogEventsOutput{
		Events: []*cloudwatchlogs.OutputLogEvent{
			{
				Message:   aws.String("[Container] 2018/01/26 18:20:01 Waiting for agent ping\n"),
				Timestamp: aws.Int64(aws.TimeUnixMilli(time.Date(2018, time.February, 3, 10, 30, 1, 0, time.FixedZone("Asia/Tokyo", 9*60*60)))),
			},
			{
				Message:   aws.String("[Container] 2018/01/26 18:20:04 Phase context status code:  Message: \n"),
				Timestamp: aws.Int64(aws.TimeUnixMilli(time.Date(2018, time.February, 3, 10, 31, 22, 0, time.FixedZone("Asia/Tokyo", 9*60*60)))),
			},
		},
	}, nil)

	return cloudWatchLogsMock
}

func mockECS(ctrl *gomock.Controller) *mock.MockECSAPI {
	ecsMock := mock.NewMockECSAPI(ctrl)

	// Mock ecs.DescribeServices
	ecsMock.EXPECT().DescribeServices(&ecs.DescribeServicesInput{
		Cluster:  aws.String("TestApp"),
		Services: []*string{aws.String("TestApp")},
	}).Return(&ecs.DescribeServicesOutput{
		Services: []*ecs.Service{
			{
				Events: []*ecs.ServiceEvent{
					{
						Id:        aws.String("354a98fa-8c77-4dc6-9c43-1ca33f293ea4"),
						CreatedAt: aws.Time(time.Date(2018, time.February, 3, 1, 35, 10, 0, time.FixedZone("UTC", 0))),
						Message:   aws.String("(service TestApp) has reached a steady state."),
					},
					{
						Id:        aws.String("5bd5b863-72e8-4f51-a255-33c7c0721345"),
						CreatedAt: aws.Time(time.Date(2018, time.February, 3, 1, 34, 56, 0, time.FixedZone("UTC", 0))),
						Message:   aws.String("(service TestApp) has stopped 1 running tasks: (task 2cf5252f-4b9e-48c3-ba73-76c1aa42e323)"),
					},
					{
						Id:        aws.String("8720a9e8-2a5a-4f83-8b01-d9fc740fa6e4"),
						CreatedAt: aws.Time(time.Date(2018, time.February, 3, 1, 32, 22, 0, time.FixedZone("UTC", 0))),
						Message:   aws.String("(service TestApp) has started 1 running tasks: (task 2cf5252f-4b9e-48c3-ba73-76c1aa42e323)"),
					},
					{
						Id:        aws.String("a990c8e1-7190-463f-af65-49446c23741c"),
						CreatedAt: aws.Time(time.Date(2018, time.February, 2, 11, 0, 5, 0, time.FixedZone("UTC", 0))),
						Message:   aws.String("(service TestApp) has reached a steady state."),
					},
				},
			},
		},
	}, nil)

	return ecsMock
}

func mockECSNotFound(ctrl *gomock.Controller) *mock.MockECSAPI {
	ecsMock := mock.NewMockECSAPI(ctrl)

	// Mock ecs.DescribeServices
	ecsMock.EXPECT().DescribeServices(&ecs.DescribeServicesInput{
		Cluster:  aws.String("TestApp"),
		Services: []*string{aws.String("TestApp")},
	}).Return(nil, awserr.New(ecs.ErrCodeClusterNotFoundException, "Not found", errors.New("Not found")))

	return ecsMock
}
