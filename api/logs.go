package api

import (
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/codebuild"
	"github.com/aws/aws-sdk-go-v2/service/codepipeline"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	log "github.com/sirupsen/logrus"
)

type Log struct {
	Message string
}

func (c *Client) DescribeLogs(appName string) []*Log {
	pipelineReq := c.Codepipeline.ListPipelineExecutionsRequest(&codepipeline.ListPipelineExecutionsInput{
		PipelineName: aws.String(appName),
	})

	pipelineResp, err := pipelineReq.Send()
	if err != nil {
		log.WithFields(log.Fields{
			"appName": appName,
		}).Fatal("Failed to get the pipeline: " + err.Error())
	}

	executions := pipelineResp.PipelineExecutionSummaries
	if len(executions) == 0 {
		return []*Log{}
	}

	var logs []*Log
	for _, stage := range c.describeStageStates(appName, *executions[0].PipelineExecutionId) {
		switch *stage.StageName {
		case "Source":
			logs = append(logs, extractSourceLogs(&stage)...)
		case "Build":
			logs = append(logs, c.describeBuildLogs(*stage.ActionStates[0].LatestExecution.ExternalExecutionId)...)
		case "Staging":
			logs = append(logs, c.describeECSServiceLogs(appName)...)
		default:
			log.Fatalf("Unexpected pipeline stages detected: %s", *stage.StageName)
		}
	}

	return logs
}

func (c *Client) describeStageStates(appName string, executionId string) []codepipeline.StageState {
	req := c.Codepipeline.GetPipelineStateRequest(&codepipeline.GetPipelineStateInput{
		Name: aws.String(appName),
	})

	resp, err := req.Send()
	if err != nil {
		log.WithFields(log.Fields{
			"appName":     appName,
			"executionId": executionId,
		}).Fatal("Failed to get the pipeline stage states: " + err.Error())
	}

	var stageStates []codepipeline.StageState
	for _, stage := range resp.StageStates {
		if *stage.LatestExecution.PipelineExecutionId == executionId {
			stageStates = append(stageStates, stage)
		}
	}

	return stageStates
}

func (c *Client) describeBuildLogs(buildId string) []*Log {
	batchGetBuildsRequest := c.Codebuild.BatchGetBuildsRequest(&codebuild.BatchGetBuildsInput{
		Ids: []string{buildId},
	})

	batchGetBuildsResponse, err := batchGetBuildsRequest.Send()
	if err != nil {
		log.WithFields(log.Fields{
			"buildId": buildId,
		}).Fatal("Failed to get the build: " + err.Error())
	}

	group := batchGetBuildsResponse.Builds[0].Logs.GroupName
	stream := batchGetBuildsResponse.Builds[0].Logs.StreamName

	getLogEventsRequest := c.Cloudwatchlogs.GetLogEventsRequest(&cloudwatchlogs.GetLogEventsInput{
		LogGroupName:  group,
		LogStreamName: stream,
	})

	getLogEventsResponse, err := getLogEventsRequest.Send()
	if err != nil {
		log.WithFields(log.Fields{
			"group":  group,
			"stream": stream,
		}).Fatal("Failed to get the build logs: " + err.Error())
	}

	var logs []*Log
	for _, event := range getLogEventsResponse.Events {
		logs = append(logs, &Log{
			Message: strings.TrimRight(*event.Message, "\n"),
		})
	}

	return logs
}

func (c *Client) describeECSServiceLogs(appName string) []*Log {
	req := c.Ecs.DescribeServicesRequest(&ecs.DescribeServicesInput{
		Cluster:  aws.String(appName),
		Services: []string{appName},
	})

	resp, err := req.Send()
	if err != nil {
		log.WithFields(log.Fields{
			"appName": appName,
		}).Fatal("Failed to get the ECS service: " + err.Error())
	}

	var logs []*Log
	for _, event := range resp.Services[0].Events {
		logs = append(logs, &Log{
			Message: *event.Message,
		})
	}

	return logs
}

func extractSourceLogs(stage *codepipeline.StageState) []*Log {
	if stage.LatestExecution.Status == "Succeeded" {
		return []*Log{
			{
				Message: "Source changes detected.",
			},
		}
	}

	return []*Log{}
}
