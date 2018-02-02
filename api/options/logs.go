package options

// DescribeLogs is the options for DescribeLogs API.
// Process is the name of a process. (e.g. builder, deployer, etc...)
// Source is the name of a log source. (e.g. app, herogate, etc...)
type DescribeLogs struct {
	Process string
	Source  string
}
