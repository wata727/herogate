package iface

import (
	"github.com/wata727/herogate/api/objects"
	"github.com/wata727/herogate/api/options"
	"github.com/wata727/herogate/log"
)

// ClientInterface is the API client's interface.
type ClientInterface interface {
	ListApps() []*objects.App
	CreateApp(appName string) *objects.App
	GetAppCreationProgress(appName string) int
	DescribeLogs(appName string, options *options.DescribeLogs) ([]*log.Log, error)
	GetApp(appName string) (*objects.App, error)
	GetTemplate(appName string) string
	DestroyApp(appName string) error
	GetAppDeletionProgress(appName string) int
	StackExists(stackName string) bool
}
