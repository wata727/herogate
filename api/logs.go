package api

import (
	"fmt"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/codebuild"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/sirupsen/logrus"
	"github.com/wata727/herogate/log"
)

// DescribeLogsOptions is the options for DescribeLogs API.
// Process is the name of a process. (e.g. builder, deployer, etc...)
// Source is the name of a log source. (e.g. app, herogate, etc...)
type DescribeLogsOptions struct {
	Process string
	Source  string
}

// DescribeLogs returns the Herogate application logs.
// In this function, it calls CodeBuild API, CloudWatchLogs API, and ECS Service API internally
// and sorts logs by timestamps.
func (c *Client) DescribeLogs(appName string, options *DescribeLogsOptions) []*log.Log {
	if options == nil {
		return []*log.Log{}
	}

	var logs []*log.Log
	switch options.Source {
	case "":
		fallthrough
	case log.HEROGATE_SOURCE:
		switch options.Process {
		case "":
			logs = append(logs, c.describeBuilderLogs(appName)...)
			logs = append(logs, c.describeDeployerLogs(appName)...)
		case log.BUIDLER_PROCESS:
			logs = append(logs, c.describeBuilderLogs(appName)...)
		case log.DEPLOYER_PROCESS:
			logs = append(logs, c.describeDeployerLogs(appName)...)
		}
	}

	sort.Slice(logs, func(i, j int) bool {
		return logs[i].Timestamp.Before(logs[j].Timestamp)
	})

	return logs
}

func (c *Client) describeBuilderLogs(appName string) []*log.Log {
	listBuildsForProjectRequest := c.CodeBuild.ListBuildsForProjectRequest(&codebuild.ListBuildsForProjectInput{
		ProjectName: aws.String(appName),
	})

	listBuildsForProjectResponse, err := listBuildsForProjectRequest.Send()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"ProjectName": appName,
		}).Fatal("Failed to get the project: " + err.Error())
	}

	if len(listBuildsForProjectResponse.Ids) == 0 {
		return []*log.Log{}
	}
	buildID := listBuildsForProjectResponse.Ids[0]
	batchGetBuildsRequest := c.CodeBuild.BatchGetBuildsRequest(&codebuild.BatchGetBuildsInput{
		Ids: []string{buildID},
	})

	batchGetBuildsResponse, err := batchGetBuildsRequest.Send()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"Build ID": buildID,
		}).Fatal("Failed to get the build: " + err.Error())
	}

	if len(batchGetBuildsResponse.Builds) == 0 {
		return []*log.Log{}
	}
	group := batchGetBuildsResponse.Builds[0].Logs.GroupName
	stream := batchGetBuildsResponse.Builds[0].Logs.StreamName

	getLogEventsRequest := c.CloudWatchLogs.GetLogEventsRequest(&cloudwatchlogs.GetLogEventsInput{
		LogGroupName:  group,
		LogStreamName: stream,
	})

	getLogEventsResponse, err := getLogEventsRequest.Send()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"LogGroupName":  group,
			"LogStreamName": stream,
		}).Fatal("Failed to get the build logs: " + err.Error())
	}

	var logs []*log.Log
	for _, event := range getLogEventsResponse.Events {
		logs = append(logs, &log.Log{
			Id:        fmt.Sprintf("%s-%d-%s", buildID, aws.Int64Value(event.Timestamp), aws.StringValue(event.Message)),
			Timestamp: aws.MillisecondsTimeValue(event.Timestamp).UTC(),
			Source:    log.HEROGATE_SOURCE,
			Process:   log.BUIDLER_PROCESS,
			Message:   strings.TrimRight(aws.StringValue(event.Message), "\n"),
		})
	}

	return logs
}

func (c *Client) describeDeployerLogs(appName string) []*log.Log {
	req := c.ECS.DescribeServicesRequest(&ecs.DescribeServicesInput{
		Cluster:  aws.String(appName),
		Services: []string{appName},
	})

	resp, err := req.Send()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"appName": appName,
		}).Fatal("Failed to get the ECS service: " + err.Error())
	}

	if len(resp.Services) == 0 {
		return []*log.Log{}
	}

	var logs []*log.Log
	for _, event := range resp.Services[0].Events {
		logs = append(logs, &log.Log{
			Id:        aws.StringValue(event.Id),
			Timestamp: aws.TimeValue(event.CreatedAt),
			Source:    log.HEROGATE_SOURCE,
			Process:   log.DEPLOYER_PROCESS,
			Message:   aws.StringValue(event.Message),
		})
	}

	return logs
}
