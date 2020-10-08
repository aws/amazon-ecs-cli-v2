// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"errors"
	"fmt"

	"github.com/aws/copilot-cli/internal/pkg/workspace"

	"github.com/aws/copilot-cli/internal/pkg/deploy"

	"github.com/spf13/cobra"

	awssession "github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/copilot-cli/internal/pkg/aws/ecr"
	"github.com/aws/copilot-cli/internal/pkg/aws/sessions"
	"github.com/aws/copilot-cli/internal/pkg/config"
	"github.com/aws/copilot-cli/internal/pkg/deploy/cloudformation"
	"github.com/aws/copilot-cli/internal/pkg/term/color"
	"github.com/aws/copilot-cli/internal/pkg/term/log"
	termprogress "github.com/aws/copilot-cli/internal/pkg/term/progress"
	"github.com/aws/copilot-cli/internal/pkg/term/prompt"
	"github.com/aws/copilot-cli/internal/pkg/term/selector"
)

const (
	fmtJobDeleteConfirmPrompt        = "Are you sure you want to delete job %s from application %s?"
	fmtJobDeleteFromEnvConfirmPrompt = "Are you sure you want to delete job %s from environment %s?"
	jobDeleteAppNamePrompt           = "Which application's job would you like to delete?"
	jobDeleteJobNamePrompt           = "Which job would you like to delete?"
	jobDeleteConfirmHelp             = "This will remove the job from all environments and delete it from your app."
	fmtJobDeleteFromEnvConfirmHelp   = "This will remove the job from just the %s environment."
)

const (
	fmtJobDeleteStart             = "Deleting job %s from environment %s."
	fmtJobDeleteFailed            = "Failed to delete job %s from environment %s: %v."
	fmtJobDeleteComplete          = "Deleted job %s from environment %s."
	fmtJobDeleteResourcesStart    = "Deleting resources of job %s from application %s."
	fmtJobDeleteResourcesComplete = "Deleted resources of job %s from application %s."
)

var (
	errJobDeleteCancelled = errors.New("job delete cancelled - no changes made")
)

type deleteJobVars struct {
	appName          string
	skipConfirmation bool
	name             string
	envName          string
}

type deleteJobOpts struct {
	deleteJobVars

	// Interfaces to dependencies.
	store     store
	prompt    prompter
	sel       wsSelector
	sess      sessionProvider
	spinner   progress
	appCFN    jobRemoverFromApp
	getJobCFN func(session *awssession.Session) wlDeleter
	getECR    func(session *awssession.Session) imageRemover
}

func newDeleteJobOpts(vars deleteJobVars) (*deleteJobOpts, error) {
	store, err := config.NewStore()
	if err != nil {
		return nil, fmt.Errorf("new config store: %w", err)
	}

	provider := sessions.NewProvider()
	defaultSession, err := provider.Default()
	if err != nil {
		return nil, err
	}
	ws, err := workspace.New()
	if err != nil {
		return nil, fmt.Errorf("new workspace: %w", err)
	}
	prompter := prompt.New()
	return &deleteJobOpts{
		deleteJobVars: vars,

		store:   store,
		spinner: termprogress.NewSpinner(),
		prompt:  prompt.New(),
		sel:     selector.NewWorkspaceSelect(prompter, store, ws),
		sess:    provider,
		appCFN:  cloudformation.New(defaultSession),
		getJobCFN: func(session *awssession.Session) wlDeleter {
			return cloudformation.New(session)
		},
		getECR: func(session *awssession.Session) imageRemover {
			return ecr.New(session)
		},
	}, nil
}

// Validate returns an error if the user inputs are invalid.
func (o *deleteJobOpts) Validate() error {
	if o.name != "" {
		if _, err := o.store.GetJob(o.appName, o.name); err != nil {
			return err
		}
	}
	if o.envName != "" {
		return o.validateEnvName()
	}
	return nil
}

// Ask prompts the user for any required flags.
func (o *deleteJobOpts) Ask() error {
	if err := o.askAppName(); err != nil {
		return err
	}
	if err := o.askJobName(); err != nil {
		return err
	}

	if o.skipConfirmation {
		return nil
	}

	// When there's no env name passed in, we'll completely
	// remove the job from the application.
	deletePrompt := fmt.Sprintf(fmtJobDeleteConfirmPrompt, o.name, o.appName)
	deleteConfirmHelp := jobDeleteConfirmHelp
	if o.envName != "" {
		// When a customer provides a particular environment,
		// we'll just delete the job from that environment -
		// but keep it in the app.
		deletePrompt = fmt.Sprintf(fmtJobDeleteFromEnvConfirmPrompt, o.name, o.envName)
		deleteConfirmHelp = fmt.Sprintf(fmtJobDeleteFromEnvConfirmHelp, o.envName)
	}

	deleteConfirmed, err := o.prompt.Confirm(
		deletePrompt,
		deleteConfirmHelp)

	if err != nil {
		return fmt.Errorf("job delete confirmation prompt: %w", err)
	}
	if !deleteConfirmed {
		return errJobDeleteCancelled
	}
	return nil
}

// Execute deletes the job's CloudFormation stack.
// If the job is being removed from the application, Execute will
// also delete the ECR repository and the SSM parameter.
func (o *deleteJobOpts) Execute() error {
	envs, err := o.appEnvironments()
	if err != nil {
		return err
	}

	if err := o.deleteStacks(envs); err != nil {
		return err
	}

	// Skip removing the job from the application if
	// we are only removing the stack from a particular environment.
	if !o.needsAppCleanup() {
		return nil
	}

	if err := o.emptyECRRepos(envs); err != nil {
		return err
	}
	if err := o.removeJobFromApp(); err != nil {
		return err
	}
	if err := o.deleteSSMParam(); err != nil {
		return err
	}

	log.Successf("Deleted job %s from application %s.\n", o.name, o.appName)

	return nil
}

func (o *deleteJobOpts) validateEnvName() error {
	if _, err := o.targetEnv(); err != nil {
		return err
	}
	return nil
}

func (o *deleteJobOpts) targetEnv() (*config.Environment, error) {
	env, err := o.store.GetEnvironment(o.appName, o.envName)
	if err != nil {
		return nil, fmt.Errorf("get environment %s from config store: %w", o.envName, err)
	}
	return env, nil
}

func (o *deleteJobOpts) askAppName() error {
	if o.appName != "" {
		return nil
	}

	name, err := o.sel.Application(jobDeleteAppNamePrompt, "")
	if err != nil {
		return fmt.Errorf("select application name: %w", err)
	}
	o.appName = name
	return nil
}

func (o *deleteJobOpts) askJobName() error {
	if o.name != "" {
		return nil
	}

	name, err := o.sel.Job(jobDeleteJobNamePrompt, "")
	if err != nil {
		return fmt.Errorf("select job: %w", err)
	}
	o.name = name
	return nil
}

func (o *deleteJobOpts) appEnvironments() ([]*config.Environment, error) {
	var envs []*config.Environment
	var err error
	if o.envName != "" {
		env, err := o.targetEnv()
		if err != nil {
			return nil, err
		}
		envs = append(envs, env)
	} else {
		envs, err = o.store.ListEnvironments(o.appName)
		if err != nil {
			return nil, fmt.Errorf("list environments: %w", err)
		}
	}
	return envs, nil
}

func (o *deleteJobOpts) deleteStacks(envs []*config.Environment) error {
	for _, env := range envs {
		sess, err := o.sess.FromRole(env.ManagerRoleARN, env.Region)
		if err != nil {
			return err
		}

		cfClient := o.getJobCFN(sess)
		o.spinner.Start(fmt.Sprintf(fmtJobDeleteStart, o.name, env.Name))
		if err := cfClient.DeleteWorkload(deploy.DeleteWorkloadInput{
			Name:    o.name,
			EnvName: env.Name,
			AppName: o.appName,
		}); err != nil {
			o.spinner.Stop(log.Serrorf(fmtJobDeleteFailed, o.name, env.Name, err))
			return err
		}
		o.spinner.Stop(log.Ssuccessf(fmtJobDeleteComplete, o.name, env.Name))
	}
	return nil
}

func (o *deleteJobOpts) needsAppCleanup() bool {
	// Only remove a job from the app if
	// we're removing it from every environment.
	// If we're just removing the job from one
	// env, we keep the app configuration.
	return o.envName == ""
}

// This is to make mocking easier in unit tests
func (o *deleteJobOpts) emptyECRRepos(envs []*config.Environment) error {
	var uniqueRegions []string
	for _, env := range envs {
		if !contains(env.Region, uniqueRegions) {
			uniqueRegions = append(uniqueRegions, env.Region)
		}
	}

	// TODO: centralized ECR repo name
	repoName := fmt.Sprintf("%s/%s", o.appName, o.name)
	for _, region := range uniqueRegions {
		sess, err := o.sess.DefaultWithRegion(region)
		if err != nil {
			return err
		}
		client := o.getECR(sess)
		if err := client.ClearRepository(repoName); err != nil {
			return err
		}
	}
	return nil
}

func (o *deleteJobOpts) removeJobFromApp() error {
	proj, err := o.store.GetApplication(o.appName)
	if err != nil {
		return err
	}

	o.spinner.Start(fmt.Sprintf(fmtJobDeleteResourcesStart, o.name, o.appName))
	if err := o.appCFN.RemoveJobFromApp(proj, o.name); err != nil {
		if !isStackSetNotExistsErr(err) {
			o.spinner.Stop(log.Serrorf(fmtJobDeleteResourcesStart, o.name, o.appName))
			return err
		}
	}
	o.spinner.Stop(log.Ssuccessf(fmtJobDeleteResourcesComplete, o.name, o.appName))
	return nil
}

func (o *deleteJobOpts) deleteSSMParam() error {
	if err := o.store.DeleteJob(o.appName, o.name); err != nil {
		return fmt.Errorf("delete job %s in application %s from config store: %w", o.name, o.appName, err)
	}

	return nil
}

// RecommendedActions returns follow-up actions the user can take after successfully executing the command.
func (o *deleteJobOpts) RecommendedActions() []string {
	return []string{
		fmt.Sprintf("Run %s to update the corresponding pipeline if it exists.",
			color.HighlightCode("copilot pipeline update")),
	}
}

// buildJobDeleteCmd builds the command to delete job(s).
func buildJobDeleteCmd() *cobra.Command {
	vars := deleteJobVars{}
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Deletes a job from an application.",
		Example: `
  Delete the "report-generator" job from the my-app application.
  /code $ copilot job delete --name report-generator --app my-app

  Delete the "report-generator" job from just the prod environment.
  /code $ copilot job delete --name report-generator --env prod

  Delete the "report-generator" job from the my-app application from outside of the workspace.
  /code $ copilot job delete --name report-generator --app my-app

  Delete the "report-generator" job without confirmation prompt.
  /code $ copilot job delete --name report-generator --yes`,
		RunE: runCmdE(func(cmd *cobra.Command, args []string) error {
			opts, err := newDeleteJobOpts(vars)
			if err != nil {
				return err
			}
			if err := opts.Validate(); err != nil {
				return err
			}
			if err := opts.Ask(); err != nil {
				return err
			}
			if err := opts.Execute(); err != nil {
				return err
			}

			log.Infoln("Recommended follow-up actions:")
			for _, followup := range opts.RecommendedActions() {
				log.Infof("- %s\n", followup)
			}
			return nil
		}),
	}

	cmd.Flags().StringVarP(&vars.appName, appFlag, appFlagShort, tryReadingAppName(), appFlagDescription)
	cmd.Flags().StringVarP(&vars.name, nameFlag, nameFlagShort, "", jobFlagDescription)
	cmd.Flags().StringVarP(&vars.envName, envFlag, envFlagShort, "", envFlagDescription)
	cmd.Flags().BoolVar(&vars.skipConfirmation, yesFlag, false, yesFlagDescription)
	return cmd
}
