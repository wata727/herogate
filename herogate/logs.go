package herogate

import (
	"fmt"
	"time"

	"github.com/urfave/cli"
	"github.com/wata727/herogate/api"
	"github.com/wata727/herogate/log"
)

func Logs(c *cli.Context) {
	client := api.NewClient()

	var lastEventLog *log.Log
	for _, eventLog := range fetchNewLogs(client, "fargateTest", c, lastEventLog) {
		lastEventLog = eventLog
		fmt.Fprintln(c.App.Writer, eventLog.Format())
	}

	for c.Bool("tail") {
		time.Sleep(5 * time.Second)
		for _, eventLog := range fetchNewLogs(client, "fargateTest", c, lastEventLog) {
			lastEventLog = eventLog
			fmt.Fprintln(c.App.Writer, eventLog.Format())
		}
	}
}

func fetchNewLogs(client *api.Client, appName string, c *cli.Context, lastEventLog *log.Log) []*log.Log {
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

	return []*log.Log{}
}
