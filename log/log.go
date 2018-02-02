package log

import (
	"fmt"
	"time"

	"github.com/fatih/color"
)

// Log is a Herogate log. This is including ID, timestamp, log source, and log process.
type Log struct {
	ID        string
	Timestamp time.Time
	Source    string
	Process   string
	Message   string
}

const (
	// HerogateSource is a kind of source type. This type occurs from Herogate internal events.
	HerogateSource = "herogate"
)

const (
	// BuilderProcess is a kind of process type. This type occurs from builder events.
	BuilderProcess = "builder"
	// DeployerProcess is a kind of process type. This type occurs from deployer events.
	DeployerProcess = "deployer"
)

// Format returns formatted text. This text including source, process, and timestamp (RFC3339).
func (l *Log) Format() string {
	herogateColor := color.New(color.FgGreen)
	timestamp := herogateColor.Sprint(l.Timestamp.Format(time.RFC3339))
	meta := herogateColor.Sprintf("%s[%s]:", l.Source, l.Process)

	return fmt.Sprintf("%s %s %s", timestamp, meta, l.Message)
}
