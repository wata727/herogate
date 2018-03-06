package herogate

import (
	"fmt"
	"io/ioutil"
	"sort"

	"github.com/hecticjeff/procfile"
	"github.com/olebedev/config"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"github.com/wata727/herogate/api"
	"github.com/wata727/herogate/api/iface"
	"github.com/wata727/herogate/container"
)

type internalGenerateTemplateContext struct {
	name     string
	image    string
	procfile string
	app      *cli.App
	client   iface.ClientInterface
}

// InternalGenerateTemplate generates new stack template from image name.
// It gets template from the current stack and replace image by specified new image name.
// Finally, it puts generated template to stdout.
func InternalGenerateTemplate(ctx *cli.Context) error {
	name := ctx.Args().First()
	if name == "" {
		return cli.NewExitError("ERROR: The application is required", 1)
	}
	image := ctx.Args().Get(1)
	if image == "" {
		return cli.NewExitError("ERROR: The image is required", 1)
	}

	file, err := ioutil.ReadFile("Procfile")
	if err != nil {
		logrus.Debug("Failed to load Procfile")
	}

	processInternalGenerateTemplate(&internalGenerateTemplateContext{
		name:     name,
		image:    image,
		procfile: string(file),
		app:      ctx.App,
		client:   api.NewClient(&api.ClientOption{}),
	})

	return nil
}

func processInternalGenerateTemplate(ctx *internalGenerateTemplateContext) {
	template := ctx.client.GetTemplate(ctx.name)
	cfg, err := config.ParseYaml(template)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"appName":  ctx.name,
			"template": template,
		}).Fatal("Failed to parse yaml template" + err.Error())
	}

	environment, err := cfg.List("Resources.HerogateApplicationContainer.Properties.ContainerDefinitions.0.Environment")
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"config": cfg,
		}).Debug("Failed to get environment list" + err.Error())
	}

	definitions := []*container.Definition{}
	proclist := procfile.Parse(ctx.procfile)
	for name, process := range proclist {
		definitions = append(
			definitions,
			container.New(name, ctx.image, append([]string{process.Command}, process.Arguments...), environment),
		)
	}
	sort.Slice(definitions, func(i, j int) bool {
		return definitions[i].Name < definitions[j].Name
	})

	if len(definitions) > 0 {
		err = cfg.Set("Resources.HerogateApplicationContainer.Properties.ContainerDefinitions", definitions)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"definitions": definitions,
				"config":      cfg,
			}).Fatal("Failed to set container definitions to template" + err.Error())
		}
	}

	result, err := config.RenderYaml(cfg.Root)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"appName": ctx.name,
			"config":  cfg.Root,
		}).Fatal("Failed to render yaml template" + err.Error())
	}

	fmt.Fprintln(ctx.app.Writer, result)
}
