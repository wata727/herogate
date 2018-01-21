package builder

import (
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"github.com/wata727/herogate/api"
)

func Logs(c *cli.Context) {
	client := api.NewClient()

	for _, eventLog := range client.DescribeBuilderLogs("fargateTest") {
		log.Info(eventLog.Message)
	}
}
