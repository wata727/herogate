package herogate

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/fatih/color"
	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
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
		Process: "ps",
		Source:  "source",
	}).Return([]*log.Log{
		{
			Id:        "foo",
			Timestamp: time.Date(2018, time.February, 2, 11, 0, 5, 0, time.FixedZone("UTC", 0)),
			Source:    "source",
			Process:   "ps",
			Message:   "foo message",
		},
		{
			Id:        "bar",
			Timestamp: time.Date(2018, time.February, 2, 11, 0, 9, 0, time.FixedZone("UTC", 0)),
			Source:    "source",
			Process:   "ps",
			Message:   "bar message",
		},
		{
			Id:        "baz",
			Timestamp: time.Date(2018, time.February, 2, 11, 0, 29, 0, time.FixedZone("UTC", 0)),
			Source:    "source",
			Process:   "ps",
			Message:   "baz message",
		},
	})

	processLogs(&logsContext{
		name:   "fargateTest",
		app:    app,
		client: client,
		num:    2,
		ps:     "ps",
		source: "source",
		tail:   false,
	})

	herogateColor := color.New(color.FgGreen)
	log1Timestamps := herogateColor.Sprint("2018-02-02T11:00:09Z")
	log2Timestamps := herogateColor.Sprint("2018-02-02T11:00:29Z")
	logMeta := herogateColor.Sprint("source[ps]:")
	expected := fmt.Sprintf(`%s %s bar message
%s %s baz message
`, log1Timestamps, logMeta, log2Timestamps, logMeta)

	if writer.String() != expected {
		t.Fatalf("\nExpected: %s\nActual: %s", expected, writer.String())
	}
}

func TestFetchNewLogs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := mock.NewMockClientInterface(ctrl)
	client.EXPECT().DescribeLogs("fargateTest", &api.DescribeLogsOptions{
		Process: "ps",
		Source:  "source",
	}).Return([]*log.Log{
		{
			Id:        "foo",
			Timestamp: time.Date(2018, time.February, 2, 11, 0, 5, 0, time.FixedZone("UTC", 0)),
			Source:    "source",
			Process:   "ps",
			Message:   "foo message",
		},
		{
			Id:        "bar",
			Timestamp: time.Date(2018, time.February, 2, 11, 0, 9, 0, time.FixedZone("UTC", 0)),
			Source:    "source",
			Process:   "ps",
			Message:   "bar message",
		},
		{
			Id:        "baz",
			Timestamp: time.Date(2018, time.February, 2, 11, 0, 29, 0, time.FixedZone("UTC", 0)),
			Source:    "source",
			Process:   "ps",
			Message:   "baz message",
		},
	})

	logs := fetchNewLogs(&logsContext{
		name:   "fargateTest",
		app:    cli.NewApp(),
		client: client,
		num:    100,
		ps:     "ps",
		source: "source",
		tail:   false,
	}, &log.Log{
		Id:        "bar",
		Timestamp: time.Date(2018, time.February, 2, 11, 0, 9, 0, time.FixedZone("UTC", 0)),
		Source:    "source",
		Process:   "ps",
		Message:   "bar message",
	})

	expected := []*log.Log{
		{
			Id:        "baz",
			Timestamp: time.Date(2018, time.February, 2, 11, 0, 29, 0, time.FixedZone("UTC", 0)),
			Source:    "source",
			Process:   "ps",
			Message:   "baz message",
		},
	}
	if !cmp.Equal(expected, logs) {
		t.Fatalf("\nDiff: %s\n", cmp.Diff(expected, logs))
	}
}
