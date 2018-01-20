package logs

import (
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/codebuild"
	"github.com/aws/aws-sdk-go-v2/service/codepipeline"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

func Logs(c *cli.Context) {
	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		panic("unable to load SDK config, " + err.Error())
	}

	svc := codepipeline.New(cfg)
	req := svc.GetPipelineStateRequest(&codepipeline.GetPipelineStateInput{
		Name: aws.String("fargate-test"),
	})

	resp, err := req.Send()
	if err != nil {
		panic("failed to API call, " + err.Error())
	}

	for _, stage := range resp.StageStates {
		switch *stage.StageName {
		case "Source":
			logsSource(&stage)
		case "Build":
			logsBuild(&stage)
		case "Staging":
		default:
			panic("Unexpected state name.")
		}
	}
}

func logsSource(s *codepipeline.StageState) {
	if s.LatestExecution.Status == "Succeeded" {
		log.Info("Source changes detected")
	}
}

func logsBuild(s *codepipeline.StageState) {
	cfg, _ := external.LoadDefaultAWSConfig()

	svc := codebuild.New(cfg)
	req := svc.BatchGetBuildsRequest(&codebuild.BatchGetBuildsInput{
		Ids: []string{*s.ActionStates[0].LatestExecution.ExternalExecutionId},
	})

	resp, err := req.Send()
	if err != nil {
		panic("failed to API call, " + err.Error())
	}

	group := resp.Builds[0].Logs.GroupName
	stream := resp.Builds[0].Logs.StreamName

	cwlSvc := cloudwatchlogs.New(cfg)
	cwlReq := cwlSvc.GetLogEventsRequest(&cloudwatchlogs.GetLogEventsInput{
		LogGroupName:  group,
		LogStreamName: stream,
	})

	cwlResp, err := cwlReq.Send()
	if err != nil {
		panic("failed to API call, " + err.Error())
	}

	for _, event := range cwlResp.Events {
		log.Info(strings.TrimRight(*event.Message, "\n"))
	}
}
