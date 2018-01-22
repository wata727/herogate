package builder

import (
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"github.com/wata727/herogate/api"
)

func Logs(c *cli.Context) {
	client := api.NewClient()

	var lastEventLog *api.Log
	for _, eventLog := range fetchNewLogs(client, "fargateTest", lastEventLog) {
		lastEventLog = eventLog
		log.Info(eventLog.Message)
	}

	for c.Bool("tail") {
		time.Sleep(5 * time.Second)
		for _, eventLog := range fetchNewLogs(client, "fargateTest", lastEventLog) {
			lastEventLog = eventLog
			log.Info(eventLog.Message)
		}
	}
}

func fetchNewLogs(client *api.Client, appName string, lastEventLog *api.Log) []*api.Log {
	eventLogs := client.DescribeBuilderLogs(appName)
	if lastEventLog == nil {
		return eventLogs
	}

	if len(eventLogs) > 0 && lastEventLog.Id != eventLogs[len(eventLogs)-1].Id {
		return eventLogs
	}

	for i := len(eventLogs) - 1; i >= 0; i-- {
		if lastEventLog.Message == eventLogs[i].Message {
			return eventLogs[i+1:]
		}
	}

	return []*api.Log{}
}
