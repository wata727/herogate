package log

import (
	"fmt"
	"testing"
	"time"

	"github.com/fatih/color"
)

func TestFormat(t *testing.T) {
	testLog := Log{
		ID:        "foo",
		Timestamp: time.Date(2018, time.February, 2, 11, 0, 9, 0, time.FixedZone("UTC", 0)),
		Source:    "source",
		Process:   "ps",
		Message:   "foo message",
	}

	herogateColor := color.New(color.FgGreen)
	logTimestamps := herogateColor.Sprint("2018-02-02T11:00:09Z")
	logMeta := herogateColor.Sprint("source[ps]:")
	expected := fmt.Sprintf("%s %s foo message", logTimestamps, logMeta)

	if testLog.Format() != expected {
		t.Fatalf("\nExpected: %s\nActual: %s", expected, testLog.Format())
	}
}
