// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/copilot-cli/cmd/copilot/template"
	awsecs "github.com/aws/copilot-cli/internal/pkg/aws/ecs"
	"github.com/aws/copilot-cli/internal/pkg/aws/sessions"
	"github.com/aws/copilot-cli/internal/pkg/cli/group"
	"github.com/aws/copilot-cli/internal/pkg/config"
	"github.com/aws/copilot-cli/internal/pkg/deploy"
	"github.com/aws/copilot-cli/internal/pkg/ecs"
	"github.com/aws/copilot-cli/internal/pkg/term/color"
	"github.com/aws/copilot-cli/internal/pkg/term/log"
	"github.com/aws/copilot-cli/internal/pkg/term/prompt"
	"github.com/aws/copilot-cli/internal/pkg/term/selector"
	"github.com/spf13/cobra"
)

const (
	svcExecNamePrompt     = "Into which service would you like to execute?"
	svcExecNameHelpPrompt = `Copilot runs your command in one of your chosen service's tasks.
The task is chosen at random, and the first essential container is used.`
)

type svcExecOpts struct {
	execVars
	store              store
	sel                deploySelector
	newSvcDescriber    func(*session.Session) serviceDescriber
	newCommandExecutor func(*session.Session) ecsCommandExecutor
	// Override in unit test
	randInt func(int) int
}

func newSvcExecOpts(vars execVars) (*svcExecOpts, error) {
	ssmStore, err := config.NewStore()
	if err != nil {
		return nil, fmt.Errorf("connect to config store: %w", err)
	}
	deployStore, err := deploy.NewStore(ssmStore)
	if err != nil {
		return nil, fmt.Errorf("connect to deploy store: %w", err)
	}
	return &svcExecOpts{
		execVars: vars,
		store:    ssmStore,
		sel:      selector.NewDeploySelect(prompt.New(), ssmStore, deployStore),
		newSvcDescriber: func(s *session.Session) serviceDescriber {
			return ecs.New(s)
		},
		newCommandExecutor: func(s *session.Session) ecsCommandExecutor {
			return awsecs.New(s)
		},
		randInt: func(x int) int {
			rand.Seed(time.Now().Unix())
			return rand.Intn(x)
		},
	}, nil
}

// Validate returns an error if the values provided by the user are invalid.
func (o *svcExecOpts) Validate() error {
	if o.appName != "" {
		if _, err := o.store.GetApplication(o.appName); err != nil {
			return err
		}
	}
	if o.envName != "" {
		if _, err := o.store.GetEnvironment(o.appName, o.envName); err != nil {
			return err
		}
	}
	if o.name != "" {
		if _, err := o.store.GetService(o.appName, o.name); err != nil {
			return err
		}
	}
	return nil
}

// Ask asks for fields that are required but not passed in.
func (o *svcExecOpts) Ask() error {
	if err := o.askApp(); err != nil {
		return err
	}
	if err := o.askSvcEnvName(); err != nil {
		return err
	}
	return nil
}

// Execute executes a command in a running container.
func (o *svcExecOpts) Execute() error {
	sess, err := o.envSession()
	if err != nil {
		return err
	}
	svcDesc, err := o.newSvcDescriber(sess).DescribeService(o.appName, o.envName, o.name)
	if err != nil {
		return fmt.Errorf("describe ECS service for %s in environment %s: %w", o.name, o.envName, err)
	}
	taskID, err := o.selectTask(awsecs.FilterRunningTasks(svcDesc.Tasks))
	if err != nil {
		return err
	}
	container := o.selectContainer()
	log.Infof("Execute into container %s in task %s.\n", container, taskID)
	execCommandErr := o.newCommandExecutor(sess).ExecuteCommand(awsecs.ExecuteCommandInput{
		Cluster:     svcDesc.ClusterName,
		Command:     o.command,
		Container:   container,
		Task:        taskID,
		Interactive: o.interactive,
	})
	if execCommandErr.TerminateErr != nil {
		log.Errorf(`Failed to terminate session %s: %s.
You can manually terminate the session by either running %s or deleting it from aws console.
`, execCommandErr.TerminateErr.SessionID,
			execCommandErr.TerminateErr.Error(),
			color.HighlightCode(fmt.Sprintf("aws ssm terminate-session --session-id %s", execCommandErr.TerminateErr.SessionID)))
	}
	if execCommandErr.Err != nil {
		return fmt.Errorf("execute command %s in container %s: %w", o.command, container, execCommandErr.Err)
	}
	return nil
}

func (o *svcExecOpts) askApp() error {
	if o.appName != "" {
		return nil
	}
	app, err := o.sel.Application(svcAppNamePrompt, svcAppNameHelpPrompt)
	if err != nil {
		return fmt.Errorf("select application: %w", err)
	}
	o.appName = app
	return nil
}

func (o *svcExecOpts) askSvcEnvName() error {
	deployedService, err := o.sel.DeployedService(svcExecNamePrompt, svcExecNameHelpPrompt, o.appName, selector.WithEnv(o.envName), selector.WithSvc(o.name))
	if err != nil {
		return fmt.Errorf("select deployed service for application %s: %w", o.appName, err)
	}
	o.name = deployedService.Svc
	o.envName = deployedService.Env
	return nil
}

func (o *svcExecOpts) envSession() (*session.Session, error) {
	env, err := o.store.GetEnvironment(o.appName, o.envName)
	if err != nil {
		return nil, fmt.Errorf("get environment %s: %w", o.envName, err)
	}
	return sessions.NewProvider().FromRole(env.ManagerRoleARN, env.Region)
}

func (o *svcExecOpts) selectTask(tasks []*awsecs.Task) (string, error) {
	if len(tasks) == 0 {
		return "", fmt.Errorf("found no running task for service %s in environment %s", o.name, o.envName)
	}
	if o.taskID != "" {
		for _, task := range tasks {
			taskID, err := awsecs.TaskID(aws.StringValue(task.TaskArn))
			if err != nil {
				return "", err
			}
			if strings.HasPrefix(taskID, o.taskID) {
				return taskID, nil
			}
		}
		return "", fmt.Errorf("found no running task whose ID is prefixed with %s", o.taskID)
	}
	taskID, err := awsecs.TaskID(aws.StringValue(tasks[o.randInt(len(tasks))].TaskArn))
	if err != nil {
		return "", err
	}
	return taskID, nil
}

func (o *svcExecOpts) selectContainer() string {
	if o.containerName != "" {
		return o.containerName
	}
	// The first essential container is named with the workload name.
	return o.name
}

// buildSvcExecCmd builds the command for execute a running container in a service.
func buildSvcExecCmd() *cobra.Command {
	vars := execVars{}
	cmd := &cobra.Command{
		Use:   "exec",
		Short: "Execute a command in a running container part of a service.",
		Example: `
  Start an interactive bash session with a task part of the "frontend" service.
  /code $ copilot svc exec -a my-app -e test -n frontend
  Runs the 'ls' command in the task prefixed with ID "8c38184" within the "backend" service.
  /code $ copilot svc exec -a my-app -e test --name backend --task-id 8c38184 --command "ls" --interactive=false`,
		RunE: runCmdE(func(cmd *cobra.Command, args []string) error {
			opts, err := newSvcExecOpts(vars)
			if err != nil {
				return err
			}
			if err := opts.Validate(); err != nil {
				return err
			}
			if err := opts.Ask(); err != nil {
				return err
			}
			return opts.Execute()
		}),
	}
	cmd.Flags().StringVarP(&vars.appName, appFlag, appFlagShort, tryReadingAppName(), appFlagDescription)
	cmd.Flags().StringVarP(&vars.envName, envFlag, envFlagShort, "", envFlagDescription)
	cmd.Flags().StringVarP(&vars.name, nameFlag, nameFlagShort, "", nameFlagDescription)
	cmd.Flags().StringVarP(&vars.command, commandFlag, commandFlagShort, defaultCommand, execCommandFlagDescription)
	cmd.Flags().StringVar(&vars.taskID, taskIDFlag, "", taskIDFlagDescription)
	cmd.Flags().StringVar(&vars.containerName, containerFlag, "", containerFlagDescription)
	cmd.Flags().BoolVar(&vars.interactive, interactiveFlag, true, interactiveFlagDescription)

	cmd.SetUsageTemplate(template.Usage)
	cmd.Annotations = map[string]string{
		"group": group.Debug,
	}
	return cmd
}
