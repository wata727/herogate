package objects

// App is Herogate application object. This is a copy of CloudFormation stack.
type App struct {
	Name            string
	Status          string
	Repository      string
	Endpoint        string
	PlatformVersion string
}

// AppInfo is Herogate application info object.
type AppInfo struct {
	*App
	Containers []*Container
	Region     string
}

// Container is Herogate application container.
type Container struct {
	Name    string
	Count   int64
	Command []string
}
