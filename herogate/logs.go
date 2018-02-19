package herogate

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"github.com/wata727/herogate/api"
	"github.com/wata727/herogate/api/iface"
	"github.com/wata727/herogate/api/options"
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

var fetchLogsInterval = 5 * time.Second

// Logs retrieves logs from builder, deployer, and app containers.
func Logs(ctx *cli.Context) error {
	region, name := detectAppFromRepo()
	if ctx.String("app") != "" {
		logrus.Debug("Override application name: " + ctx.String("app"))
		name = ctx.String("app")
	}

	return processLogs(&logsContext{
		name:   name,
		app:    ctx.App,
		client: api.NewClient(&api.ClientOption{Region: region}),
		num:    ctx.Int("num"),
		ps:     ctx.String("ps"),
		source: ctx.String("source"),
		tail:   ctx.Bool("tail"),
	})
}

func processLogs(ctx *logsContext) error {
	var lastEventLog *log.Log
	eventLogs, err := fetchNewLogs(ctx, lastEventLog)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("ERROR: Application not found: %s", ctx.name), 1)
	}
	if len(eventLogs)-ctx.num > 0 {
		eventLogs = eventLogs[len(eventLogs)-ctx.num:]
	}

	for _, eventLog := range eventLogs {
		lastEventLog = eventLog
		fmt.Fprintln(ctx.app.Writer, eventLog.Format())
	}

	for ctx.tail {
		time.Sleep(fetchLogsInterval)
		newLogs, err := fetchNewLogs(ctx, lastEventLog)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"appName": ctx.name,
			}).Fatal("Unexpected fetch error occurred: " + err.Error())
		}

		for _, eventLog := range newLogs {
			lastEventLog = eventLog
			fmt.Fprintln(ctx.app.Writer, eventLog.Format())
		}
	}

	return nil
}

func fetchNewLogs(ctx *logsContext, lastEventLog *log.Log) ([]*log.Log, error) {
	eventLogs, err := ctx.client.DescribeLogs(ctx.name, &options.DescribeLogs{
		Process: ctx.ps,
		Source:  ctx.source,
	})
	if err != nil {
		return []*log.Log{}, err
	}

	// When fetching at first, returns all logs
	if lastEventLog == nil {
		return eventLogs, nil
	}

	// When fetching for the same build, returns new logs based on log ID
	for i := len(eventLogs) - 1; i >= 0; i-- {
		if lastEventLog.ID == eventLogs[i].ID {
			return eventLogs[i+1:], nil
		}
	}

	// When fetching for other builds, returns new logs based on timestamp
	var latestLogs []*log.Log
	for _, eventLog := range eventLogs {
		if eventLog.Timestamp.After(lastEventLog.Timestamp) {
			latestLogs = append(latestLogs, eventLog)
		}
	}
	return latestLogs, nil
}
