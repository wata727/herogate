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
	"github.com/wata727/herogate/api/options"
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
	client.EXPECT().DescribeLogs("fargateTest", &options.DescribeLogs{
		Process: "ps",
		Source:  "source",
	}).Return([]*log.Log{
		{
			ID:        "foo",
			Timestamp: time.Date(2018, time.February, 2, 11, 0, 5, 0, time.FixedZone("UTC", 0)),
			Source:    "source",
			Process:   "ps",
			Message:   "foo message",
		},
		{
			ID:        "bar",
			Timestamp: time.Date(2018, time.February, 2, 11, 0, 9, 0, time.FixedZone("UTC", 0)),
			Source:    "source",
			Process:   "ps",
			Message:   "bar message",
		},
		{
			ID:        "baz",
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

	newLogs := []*log.Log{
		{
			ID:        "foo",
			Timestamp: time.Date(2018, time.February, 2, 11, 0, 5, 0, time.FixedZone("UTC", 0)),
			Source:    "source",
			Process:   "ps",
			Message:   "foo message",
		},
		{
			ID:        "bar",
			Timestamp: time.Date(2018, time.February, 2, 11, 0, 9, 0, time.FixedZone("UTC", 0)),
			Source:    "source",
			Process:   "ps",
			Message:   "bar message",
		},
		{
			ID:        "baz",
			Timestamp: time.Date(2018, time.February, 2, 11, 0, 29, 0, time.FixedZone("UTC", 0)),
			Source:    "source",
			Process:   "ps",
			Message:   "baz message",
		},
	}
	client := mock.NewMockClientInterface(ctrl)
	client.EXPECT().DescribeLogs("fargateTest", &options.DescribeLogs{
		Process: "ps",
		Source:  "source",
	}).Return(newLogs)

	logs := fetchNewLogs(&logsContext{
		name:   "fargateTest",
		app:    cli.NewApp(),
		client: client,
		num:    100,
		ps:     "ps",
		source: "source",
		tail:   false,
	}, nil)

	if !cmp.Equal(newLogs, logs) {
		t.Fatalf("\nDiff: %s\n", cmp.Diff(newLogs, logs))
	}
}

func TestFetchNewLogs__fetchingForTheSameBuild(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := mock.NewMockClientInterface(ctrl)
	client.EXPECT().DescribeLogs("fargateTest", &options.DescribeLogs{
		Process: "ps",
		Source:  "source",
	}).Return([]*log.Log{
		{
			ID:        "foo",
			Timestamp: time.Date(2018, time.February, 2, 11, 0, 5, 0, time.FixedZone("UTC", 0)),
			Source:    "source",
			Process:   "ps",
			Message:   "foo message",
		},
		{
			ID:        "bar",
			Timestamp: time.Date(2018, time.February, 2, 11, 0, 9, 0, time.FixedZone("UTC", 0)),
			Source:    "source",
			Process:   "ps",
			Message:   "bar message",
		},
		{
			ID:        "baz",
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
		ID:        "bar",
		Timestamp: time.Date(2018, time.February, 2, 11, 0, 9, 0, time.FixedZone("UTC", 0)),
		Source:    "source",
		Process:   "ps",
		Message:   "bar message",
	})

	expected := []*log.Log{
		{
			ID:        "baz",
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

func TestFetchNewLogs__fetchingForOtherBuilds(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := mock.NewMockClientInterface(ctrl)
	client.EXPECT().DescribeLogs("fargateTest", &options.DescribeLogs{
		Process: "ps",
		Source:  "source",
	}).Return([]*log.Log{
		{
			ID:        "foo",
			Timestamp: time.Date(2018, time.February, 2, 11, 0, 5, 0, time.FixedZone("UTC", 0)),
			Source:    "source",
			Process:   "ps",
			Message:   "foo message",
		},
		{
			ID:        "bar",
			Timestamp: time.Date(2018, time.February, 2, 11, 0, 9, 0, time.FixedZone("UTC", 0)),
			Source:    "source",
			Process:   "ps",
			Message:   "bar message",
		},
		{
			ID:        "baz",
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
		ID:        "hoge",
		Timestamp: time.Date(2018, time.February, 2, 11, 0, 11, 0, time.FixedZone("UTC", 0)),
		Source:    "source",
		Process:   "ps",
		Message:   "hoge message",
	})

	expected := []*log.Log{
		{
			ID:        "baz",
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
