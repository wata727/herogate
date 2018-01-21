package herogate

import (
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"github.com/wata727/herogate/api"
)

func Logs(c *cli.Context) {
	client := api.NewClient()

	for _, eventLog := range client.DescribeLogs("fargateTest") {
		log.Info(eventLog.Message)
	}
}
