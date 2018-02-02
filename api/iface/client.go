package iface

import (
	"github.com/wata727/herogate/api/options"
	"github.com/wata727/herogate/log"
)

// ClientInterface is the API client's interface.
type ClientInterface interface {
	DescribeLogs(appName string, options *options.DescribeLogs) []*log.Log
}
