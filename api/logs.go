package api

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/codebuild"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	log "github.com/sirupsen/logrus"
)

type Log struct {
	Id        string
	Timestamp time.Time
	Source    string
	Process   string
	Message   string
}

const (
	HEROGATE_SOURCE = "herogate"
)

const (
	BUIDLER_PROCESS  = "builder"
	DEPLOYER_PROCESS = "deployer"
)

type DescribeLogsOptions struct {
	Process string
	Source  string
}

func (c *Client) DescribeLogs(appName string, options *DescribeLogsOptions) []*Log {
	var logs []*Log

	if options == nil {
		return logs
	}

	if options.Source == "" || options.Source == HEROGATE_SOURCE || options.Process == "" || options.Process == BUIDLER_PROCESS {
		logs = append(logs, c.DescribeBuilderLogs(appName)...)
	}

	if options.Source == "" || options.Source == HEROGATE_SOURCE || options.Process == "" || options.Process == DEPLOYER_PROCESS {
		logs = append(logs, c.DescribeDeployerLogs(appName)...)
	}

	sort.Slice(logs, func(i, j int) bool {
		return logs[i].Timestamp.Before(logs[j].Timestamp)
	})

	return logs
}

func (c *Client) DescribeBuilderLogs(appName string) []*Log {
	listBuildsForProjectRequest := c.CodeBuild.ListBuildsForProjectRequest(&codebuild.ListBuildsForProjectInput{
		ProjectName: aws.String(appName),
	})

	listBuildsForProjectResponse, err := listBuildsForProjectRequest.Send()
	if err != nil {
		log.WithFields(log.Fields{
			"ProjectName": appName,
		}).Fatal("Failed to get the project: " + err.Error())
	}

	if len(listBuildsForProjectResponse.Ids) == 0 {
		return []*Log{}
	}
	buildId := listBuildsForProjectResponse.Ids[0]
	batchGetBuildsRequest := c.CodeBuild.BatchGetBuildsRequest(&codebuild.BatchGetBuildsInput{
		Ids: []string{buildId},
	})

	batchGetBuildsResponse, err := batchGetBuildsRequest.Send()
	if err != nil {
		log.WithFields(log.Fields{
			"BuildId": buildId,
		}).Fatal("Failed to get the build: " + err.Error())
	}

	group := batchGetBuildsResponse.Builds[0].Logs.GroupName
	stream := batchGetBuildsResponse.Builds[0].Logs.StreamName

	getLogEventsRequest := c.CloudWatchLogs.GetLogEventsRequest(&cloudwatchlogs.GetLogEventsInput{
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
			Id:        fmt.Sprintf("%s-%d-%s", buildId, aws.Int64Value(event.Timestamp), aws.StringValue(event.Message)),
			Timestamp: aws.MillisecondsTimeValue(event.Timestamp),
			Source:    HEROGATE_SOURCE,
			Process:   BUIDLER_PROCESS,
			Message:   strings.TrimRight(aws.StringValue(event.Message), "\n"),
		})
	}

	return logs
}

func (c *Client) DescribeDeployerLogs(appName string) []*Log {
	req := c.ECS.DescribeServicesRequest(&ecs.DescribeServicesInput{
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
			Id:        aws.StringValue(event.Id),
			Timestamp: aws.TimeValue(event.CreatedAt),
			Source:    HEROGATE_SOURCE,
			Process:   DEPLOYER_PROCESS,
			Message:   aws.StringValue(event.Message),
		})
	}

	return logs
}
