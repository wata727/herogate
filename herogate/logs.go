package herogate

import (
	"fmt"
	"time"

	"github.com/urfave/cli"
	"github.com/wata727/herogate/api"
	"github.com/wata727/herogate/api/iface"
	"github.com/wata727/herogate/log"
)

type logsContext struct {
	name   string
	app    *cli.App
	client iface.ClientInterface
	num    int
	ps     string
	source string
	tail   bool
}

// Logs retrieves logs from builder, deployer, and app containers.
func Logs(c *cli.Context) {
	processLogs(&logsContext{
		name:   "fargateTest",
		app:    c.App,
		client: api.NewClient(),
		num:    c.Int("num"),
		ps:     c.String("ps"),
		source: c.String("source"),
		tail:   c.Bool("tail"),
	})
}

func processLogs(ctx *logsContext) {
	var lastEventLog *log.Log
	eventLogs := fetchNewLogs(ctx, lastEventLog)
	if len(eventLogs)-ctx.num > 0 {
		eventLogs = eventLogs[len(eventLogs)-ctx.num:]
	}

	for _, eventLog := range eventLogs {
		lastEventLog = eventLog
		fmt.Fprintln(ctx.app.Writer, eventLog.Format())
	}

	for ctx.tail {
		time.Sleep(5 * time.Second)
		for _, eventLog := range fetchNewLogs(ctx, lastEventLog) {
			lastEventLog = eventLog
			fmt.Fprintln(ctx.app.Writer, eventLog.Format())
		}
	}
}

func fetchNewLogs(ctx *logsContext, lastEventLog *log.Log) []*log.Log {
	eventLogs := ctx.client.DescribeLogs(ctx.name, &api.DescribeLogsOptions{
		Process: ctx.ps,
		Source:  ctx.source,
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
