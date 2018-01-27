package api

import (
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/cloudwatchlogsiface"
	"github.com/aws/aws-sdk-go-v2/service/codebuild"
	"github.com/aws/aws-sdk-go-v2/service/codebuild/codebuildiface"
	"github.com/aws/aws-sdk-go-v2/service/codepipeline"
	"github.com/aws/aws-sdk-go-v2/service/codepipeline/codepipelineiface"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/ecsiface"
	log "github.com/sirupsen/logrus"
)

// Client is the Herogate API client.
// This is a wrapper of AWS API clients.
type Client struct {
	CodePipeline   codepipelineiface.CodePipelineAPI
	CodeBuild      codebuildiface.CodeBuildAPI
	CloudWatchLogs cloudwatchlogsiface.CloudWatchLogsAPI
	ECS            ecsiface.ECSAPI
}

// NewClient initializes a new client from AWS config.
func NewClient() *Client {
	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		log.Fatal("Unable to load SDK config: " + err.Error())
	}

	return &Client{
		CodePipeline:   codepipeline.New(cfg),
		CodeBuild:      codebuild.New(cfg),
		CloudWatchLogs: cloudwatchlogs.New(cfg),
		ECS:            ecs.New(cfg),
	}
}
