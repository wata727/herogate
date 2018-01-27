package api

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/codebuild"
	"github.com/golang/mock/gomock"
	"github.com/wata727/herogate/log"
	"github.com/wata727/herogate/mock"
)

func TestDescribeLogs(t *testing.T) {
	cases := []struct {
		Name    string
		Options *DescribeLogsOptions
		Result  []*log.Log
	}{
		{
			Name:    "No config",
			Options: &DescribeLogsOptions{},
			Result:  []*log.Log{},
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	codeBuildMock := mock.NewMockCodeBuildAPI()
	codeBuildMock.EXPECT().ListBuildsForProjectRequest(&codebuild.ListBuildsForProjectInput{
		ProjectName: aws.String("TestApp"),
	}).Return(1) // breaking...

	client := NewClient()
	client.CodeBuild = codeBuildMock

	for _, tc := range cases {
		client.DescribeLogs("TestApp", tc.Options)
	}
}
