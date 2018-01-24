package api

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/codebuild"
	"github.com/aws/aws-sdk-go-v2/service/codepipeline"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	log "github.com/sirupsen/logrus"
)

type Log struct {
	Id        string
	CreatedAt time.Time
	Message   string
}

func (c *Client) DescribeBuilderLogs(appName string) []*Log {
	executionId, err := c.describeLatestExecutionId(appName)
	if err != nil {
		return []*Log{}
	}

	var buildId string
	for _, stage := range c.describeStageStates(appName, executionId) {
		if *stage.StageName == "Build" && len(stage.ActionStates) > 0 && stage.ActionStates[0].LatestExecution.ExternalExecutionId != nil {
			buildId = *stage.ActionStates[0].LatestExecution.ExternalExecutionId
		}
	}

	if buildId == "" {
		return []*Log{}
	}

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
			Id:      buildId,
			Message: strings.TrimRight(*event.Message, "\n"),
		})
	}

	return logs
}

func (c *Client) DescribeDeployerLogs(appName string) []*Log {
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
			Id:        *event.Id,
			CreatedAt: *event.CreatedAt,
			Message:   fmt.Sprintf("%s %s", *event.CreatedAt, *event.Message),
		})
	}

	sort.Slice(logs, func(i, j int) bool {
		return logs[i].CreatedAt.Before(logs[j].CreatedAt)
	})

	return logs
}

func (c *Client) describeLatestExecutionId(appName string) (string, error) {
	req := c.Codepipeline.ListPipelineExecutionsRequest(&codepipeline.ListPipelineExecutionsInput{
		PipelineName: aws.String(appName),
	})

	resp, err := req.Send()
	if err != nil {
		log.WithFields(log.Fields{
			"appName": appName,
		}).Fatal("Failed to get the pipeline executions: " + err.Error())
	}

	executions := resp.PipelineExecutionSummaries
	if len(executions) == 0 {
		return "", errors.New("Empty executions")
	}

	return *executions[0].PipelineExecutionId, nil
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
