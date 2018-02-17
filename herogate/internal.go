package herogate

import (
	"fmt"

	"github.com/olebedev/config"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"github.com/wata727/herogate/api"
	"github.com/wata727/herogate/api/iface"
)

type internalGenerateTemplateContext struct {
	name   string
	image  string
	app    *cli.App
	client iface.ClientInterface
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

	processInternalGenerateTemplate(&internalGenerateTemplateContext{
		name:   name,
		image:  image,
		app:    ctx.App,
		client: api.NewClient(&api.ClientOption{}),
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

	err = cfg.Set("Resources.HerogateApplicationContainer.Properties.ContainerDefinitions.0.Image", ctx.image)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"appName": ctx.name,
			"config":  cfg,
		}).Fatal("Failed to set image to template" + err.Error())
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
