package herogate

import (
	"bytes"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/urfave/cli"
	"github.com/wata727/herogate/api"
	"github.com/wata727/herogate/log"
	"github.com/wata727/herogate/mock"
)

func TestProcessLogs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	app := cli.NewApp()
	writer := new(bytes.Buffer)
	app.Writer = writer

	client := mock.NewMockClientInterface(ctrl)
	client.EXPECT().DescribeLogs("fargateTest", &api.DescribeLogsOptions{
		Process: "",
		Source:  "",
	}).Return([]*log.Log{})

	processLogs(&logsService{
		name:   "fargateTest",
		app:    app,
		client: client,
		num:    100,
		ps:     "",
		source: "",
		tail:   false,
	})

	expected := ""
	if writer.String() != expected {
		t.Fatalf("\nExpected: %s\nActual: %s", expected, writer.String())
	}
}
