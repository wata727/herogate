package herogate

import (
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"github.com/wata727/herogate/api"
)

func Logs(c *cli.Context) {
	client := api.NewClient()

	var lastEventLog *api.Log
	for _, eventLog := range fetchNewLogs(client, "fargateTest", c, lastEventLog) {
		lastEventLog = eventLog
		log.Info(eventLog.Message)
	}

	for c.Bool("tail") {
		time.Sleep(5 * time.Second)
		for _, eventLog := range fetchNewLogs(client, "fargateTest", c, lastEventLog) {
			lastEventLog = eventLog
			log.Info(eventLog.Message)
		}
	}
}

func fetchNewLogs(client *api.Client, appName string, c *cli.Context, lastEventLog *api.Log) []*api.Log {
	eventLogs := client.DescribeLogs(appName, &api.DescribeLogsOptions{
		Process: c.String("process"),
		Source:  c.String("source"),
	})
	if lastEventLog == nil {
		return eventLogs
	}

	for i := len(eventLogs) - 1; i >= 0; i-- {
		if lastEventLog.Id == eventLogs[i].Id {
			return eventLogs[i+1:]
		}
	}

	return []*api.Log{}
}
