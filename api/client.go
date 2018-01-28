package api

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs/cloudwatchlogsiface"
	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/aws/aws-sdk-go/service/codebuild/codebuildiface"
	"github.com/aws/aws-sdk-go/service/codepipeline"
	"github.com/aws/aws-sdk-go/service/codepipeline/codepipelineiface"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/ecs/ecsiface"
)

// Client is the Herogate API client.
// This is a wrapper of AWS API clients.
type Client struct {
	codePipeline   codepipelineiface.CodePipelineAPI
	codeBuild      codebuildiface.CodeBuildAPI
	cloudWatchLogs cloudwatchlogsiface.CloudWatchLogsAPI
	ecs            ecsiface.ECSAPI
}

// NewClient initializes a new client from AWS config.
func NewClient() *Client {
	s := session.New()

	return &Client{
		codePipeline:   codepipeline.New(s),
		codeBuild:      codebuild.New(s),
		cloudWatchLogs: cloudwatchlogs.New(s),
		ecs:            ecs.New(s),
	}
}
