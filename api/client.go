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

type Client struct {
	Codepipeline   codepipelineiface.CodePipelineAPI
	Codebuild      codebuildiface.CodeBuildAPI
	Cloudwatchlogs cloudwatchlogsiface.CloudWatchLogsAPI
	Ecs            ecsiface.ECSAPI
}

func NewClient() *Client {
	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		log.Fatal("Unable to load SDK config: " + err.Error())
	}

	return &Client{
		Codepipeline:   codepipeline.New(cfg),
		Codebuild:      codebuild.New(cfg),
		Cloudwatchlogs: cloudwatchlogs.New(cfg),
		Ecs:            ecs.New(cfg),
	}
}