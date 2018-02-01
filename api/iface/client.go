package iface

import (
	"github.com/wata727/herogate/api"
	"github.com/wata727/herogate/log"
)

// ClientInterface is the API client's interface.
type ClientInterface interface {
	DescribeLogs(appName string, options *api.DescribeLogsOptions) []*log.Log
}
