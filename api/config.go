package api

import (
	"sort"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/olebedev/config"
	"github.com/sirupsen/logrus"
)

// DescribeEnvVars describes environment variables from container deifinitions.
func (c *Client) DescribeEnvVars(appName string) (map[string]string, error) {
	serviceResp, err := c.ecs.DescribeServices(&ecs.DescribeServicesInput{
		Cluster:  aws.String(appName),
		Services: []*string{aws.String(appName)},
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == ecs.ErrCodeClusterNotFoundException {
			return map[string]string{}, err
		}
		logrus.WithFields(logrus.Fields{
			"appName": appName,
		}).Fatal("Failed to get the ECS service: " + err.Error())
	}
	if len(serviceResp.Services) == 0 {
		logrus.WithFields(logrus.Fields{
			"appName": appName,
		}).Fatal("ECS services are not found: " + err.Error())
	}
	service := serviceResp.Services[0]

	taskResp, err := c.ecs.DescribeTaskDefinition(&ecs.DescribeTaskDefinitionInput{
		TaskDefinition: service.TaskDefinition,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"appName":  appName,
			"taskName": aws.StringValue(service.TaskDefinition),
		}).Fatal("Failed to get the ECS task definiation: " + err.Error())
	}
	if len(taskResp.TaskDefinition.ContainerDefinitions) == 0 {
		return map[string]string{}, nil
	}
	definiation := taskResp.TaskDefinition.ContainerDefinitions[0]

	envVars := map[string]string{}
	for _, env := range definiation.Environment {
		envVars[aws.StringValue(env.Name)] = aws.StringValue(env.Value)
	}

	return envVars, nil
}

// SetEnvVars updates CloudFormation stack with new environment variables.
// It generates new template by adding or merging environment variables from existing template.
// Because this operation restarts existing containers, It takes time to complete.
func (c *Client) SetEnvVars(appName string, envVars map[string]string) error {
	if _, err := c.GetApp(appName); err != nil {
		return err
	}

	template := generateUpdatedEnvVarsTemplate(c.GetTemplate(appName), envVars)

	_, err := c.cloudFormation.UpdateStack(&cloudformation.UpdateStackInput{
		StackName:    aws.String(appName),
		TemplateBody: aws.String(template),
		Capabilities: []*string{aws.String("CAPABILITY_NAMED_IAM")},
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"appName": appName,
		}).Fatal("Failed to request for updating stack: " + err.Error())
	}
	err = c.cloudFormation.WaitUntilStackUpdateComplete(&cloudformation.DescribeStacksInput{
		StackName: aws.String(appName),
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"appName": appName,
		}).Fatal("Failed to wait stack update: " + err.Error())
	}

	return nil
}

func generateUpdatedEnvVarsTemplate(base string, envVars map[string]string) string {
	cfg, err := config.ParseYaml(base)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"template": base,
		}).Fatal("Failed to parse yaml template" + err.Error())
	}

	envMap := map[string]string{}
	if envs, err := cfg.List("Resources.HerogateApplicationContainer.Properties.ContainerDefinitions.0.Environment"); err == nil {
		for _, env := range envs {
			e, ok := env.(map[string]interface{})
			if !ok {
				logrus.WithFields(logrus.Fields{
					"env": env,
				}).Fatal("Failed to cast environment")
			}

			var key string
			var value string
			for k, v := range e {
				switch k {
				case "Name":
					key, ok = v.(string)
					if !ok {
						logrus.WithFields(logrus.Fields{
							"env": env,
						}).Fatal("Failed to cast environment key")
					}
				case "Value":
					value, ok = v.(string)
					if !ok {
						logrus.WithFields(logrus.Fields{
							"env": env,
						}).Fatal("Failed to cast environment value")
					}
				}
			}
			envMap[key] = value
		}
	}

	for k, v := range envVars {
		envMap[k] = v
	}

	envList := []map[string]string{}
	for k, v := range envMap {
		envList = append(envList, map[string]string{"Name": k, "Value": v})
	}

	// Sort alphabetically
	sort.Slice(envList, func(i, j int) bool {
		return envList[i]["Name"] < envList[j]["Name"]
	})

	err = cfg.Set("Resources.HerogateApplicationContainer.Properties.ContainerDefinitions.0.Environment", envList)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"config":  cfg,
			"envList": envList,
		}).Fatal("Failed to set environments to template" + err.Error())
	}

	result, err := config.RenderYaml(cfg.Root)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"config": cfg.Root,
		}).Fatal("Failed to render yaml template" + err.Error())
	}

	return result
}

// UnsetEnvVars updates CloudFormation stack with new environment variables.
// It generates new template by deleting environment variables from existing template.
// Because this operation restarts existing containers, It takes time to complete.
func (c *Client) UnsetEnvVars(appName string, envList []string) error {
	if _, err := c.GetApp(appName); err != nil {
		return err
	}

	template := generateUnsettedEnvVarsTemplate(c.GetTemplate(appName), envList)

	_, err := c.cloudFormation.UpdateStack(&cloudformation.UpdateStackInput{
		StackName:    aws.String(appName),
		TemplateBody: aws.String(template),
		Capabilities: []*string{aws.String("CAPABILITY_NAMED_IAM")},
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"appName": appName,
		}).Fatal("Failed to request for updating stack: " + err.Error())
	}
	err = c.cloudFormation.WaitUntilStackUpdateComplete(&cloudformation.DescribeStacksInput{
		StackName: aws.String(appName),
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"appName": appName,
		}).Fatal("Failed to wait stack update: " + err.Error())
	}

	return nil
}

func generateUnsettedEnvVarsTemplate(base string, envList []string) string {
	cfg, err := config.ParseYaml(base)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"template": base,
		}).Fatal("Failed to parse yaml template" + err.Error())
	}

	envs := []map[string]interface{}{}
	if environments, err := cfg.List("Resources.HerogateApplicationContainer.Properties.ContainerDefinitions.0.Environment"); err == nil {
		for _, environment := range environments {
			env, ok := environment.(map[string]interface{})
			if !ok {
				logrus.WithFields(logrus.Fields{
					"environment": environment,
				}).Fatal("Failed to cast environment")
			}

			var ignore bool
			for key, value := range env {
				if key != "Name" {
					continue
				}
				v, ok := value.(string)
				if !ok {
					logrus.WithFields(logrus.Fields{
						"key":   key,
						"value": value,
					}).Fatal("Failed to cast name value")
				}

				for _, name := range envList {
					if v == name {
						ignore = true
					}
				}
			}

			if !ignore {
				envs = append(envs, env)
			}
		}
	}

	err = cfg.Set("Resources.HerogateApplicationContainer.Properties.ContainerDefinitions.0.Environment", envs)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"config": cfg,
			"envs":   envs,
		}).Fatal("Failed to set environments to template" + err.Error())
	}

	result, err := config.RenderYaml(cfg.Root)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"config": cfg.Root,
		}).Fatal("Failed to render yaml template" + err.Error())
	}

	return result
}
