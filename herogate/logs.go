package herogate

import (
	"fmt"
	"time"

	"github.com/urfave/cli"
	"github.com/wata727/herogate/api"
	"github.com/wata727/herogate/api/iface"
	"github.com/wata727/herogate/log"
)

type logsService struct {
	name   string
	app    *cli.App
	client iface.ClientInterface
	num    int
	ps     string
	source string
	tail   bool
}

func Logs(c *cli.Context) {
	processLogs(&logsService{
		name:   "fargateTest",
		app:    c.App,
		client: api.NewClient(),
		num:    c.Int("num"),
		ps:     c.String("ps"),
		source: c.String("source"),
		tail:   c.Bool("tail"),
	})
}

func processLogs(svc *logsService) {
	var lastEventLog *log.Log
	eventLogs := fetchNewLogs(svc, lastEventLog)
	if len(eventLogs)-svc.num > 0 {
		eventLogs = eventLogs[len(eventLogs)-svc.num:]
	}

	for _, eventLog := range eventLogs {
		lastEventLog = eventLog
		fmt.Fprintln(svc.app.Writer, eventLog.Format())
	}

	for svc.tail {
		time.Sleep(5 * time.Second)
		for _, eventLog := range fetchNewLogs(svc, lastEventLog) {
			lastEventLog = eventLog
			fmt.Fprintln(svc.app.Writer, eventLog.Format())
		}
	}
}

func fetchNewLogs(svc *logsService, lastEventLog *log.Log) []*log.Log {
	eventLogs := svc.client.DescribeLogs(svc.name, &api.DescribeLogsOptions{
		Process: svc.ps,
		Source:  svc.source,
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
