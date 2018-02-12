package objects

// App is Herogate application object. This is a copy of CloudFormation stack.
type App struct {
	Name       string
	Status     string
	Repository string
	Endpoint   string
}
