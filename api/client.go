package api

import (
	"github.com/aws/aws-sdk-go/aws"
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

//go:generate mockgen -source iface/client.go -destination ../mock/client.go -package mock
//go:generate mockgen -source ../vendor/github.com/aws/aws-sdk-go/service/codebuild/codebuildiface/interface.go -destination ../mock/codebuild.go -package mock
//go:generate mockgen -source ../vendor/github.com/aws/aws-sdk-go/service/cloudwatchlogs/cloudwatchlogsiface/interface.go -destination ../mock/cloudwatchlogs.go -package mock
//go:generate mockgen -source ../vendor/github.com/aws/aws-sdk-go/service/ecs/ecsiface/interface.go -destination ../mock/ecs.go -package mock

// Client is the Herogate API client.
// This is a wrapper of AWS API clients.
type Client struct {
	codePipeline   codepipelineiface.CodePipelineAPI
	codeBuild      codebuildiface.CodeBuildAPI
	cloudWatchLogs cloudwatchlogsiface.CloudWatchLogsAPI
	ecs            ecsiface.ECSAPI
}

// ClientOption is options for Herogate API Client.
// Regions can specify regions used in AWS.
type ClientOption struct {
	Region string
}

// NewClient initializes a new client from AWS config.
func NewClient(option *ClientOption) *Client {
	s := session.New()
	if option.Region != "" {
		s = session.New(&aws.Config{Region: aws.String(option.Region)})
	}

	return &Client{
		codePipeline:   codepipeline.New(s),
		codeBuild:      codebuild.New(s),
		cloudWatchLogs: cloudwatchlogs.New(s),
		ecs:            ecs.New(s),
	}
}
