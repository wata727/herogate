package log

import (
	"fmt"
	"time"

	"github.com/fatih/color"
)

type Log struct {
	Id        string
	Timestamp time.Time
	Source    string
	Process   string
	Message   string
}

const (
	HEROGATE_SOURCE = "herogate"
)

const (
	BUIDLER_PROCESS  = "builder"
	DEPLOYER_PROCESS = "deployer"
)

func (l *Log) Format() string {
	herogateColor := color.New(color.FgGreen)
	timestamp := herogateColor.Sprint(l.Timestamp.Format(time.RFC3339))
	meta := herogateColor.Sprintf("%s[%s]:", l.Source, l.Process)

	return fmt.Sprintf("%s %s %s", timestamp, meta, l.Message)
}
