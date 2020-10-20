// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"errors"
	"fmt"

	"github.com/aws/copilot-cli/internal/pkg/aws/sessions"
	"github.com/aws/copilot-cli/internal/pkg/cli/group"
	"github.com/aws/copilot-cli/internal/pkg/config"
	"github.com/aws/copilot-cli/internal/pkg/deploy/cloudformation"
	"github.com/aws/copilot-cli/internal/pkg/manifest"
	"github.com/aws/copilot-cli/internal/pkg/term/color"
	"github.com/aws/copilot-cli/internal/pkg/term/log"
	termprogress "github.com/aws/copilot-cli/internal/pkg/term/progress"
	"github.com/aws/copilot-cli/internal/pkg/term/prompt"
	"github.com/aws/copilot-cli/internal/pkg/term/selector"
	"github.com/aws/copilot-cli/internal/pkg/workspace"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

var (
	jobInitSchedulePrompt = "How would you like to " + color.Emphasize("schedule") + " this job?"
	jobInitScheduleHelp   = `How to determine this job's schedule. "Rate" lets you define the time between 
executions and is good for jobs which need to run frequently. "Fixed Schedule"
lets you use a predefined or custom cron schedule and is good for less-frequent 
jobs or those which require specific execution schedules.`
)

const (
	fmtAddJobToAppStart    = "Creating ECR repositories for job %s."
	fmtAddJobToAppFailed   = "Failed to create ECR repositories for job %s.\n"
	fmtAddJobToAppComplete = "Created ECR repositories for job %s.\n"
)

const (
	job = "job"
)

type initJobVars struct {
	appName        string
	name           string
	dockerfilePath string
	image          string
	timeout        string
	retries        int
	schedule       string
	jobType        string
}

type initJobOpts struct {
	initJobVars

	// Interfaces to interact with dependencies.
	fs          afero.Fs
	ws          jobDirManifestWriter
	store       store
	appDeployer appDeployer
	prog        progress
	prompt      prompter
	sel         initJobSelector

	// Outputs stored on successful actions.
	manifestPath string
}

func newInitJobOpts(vars initJobVars) (*initJobOpts, error) {
	store, err := config.NewStore()
	if err != nil {
		return nil, fmt.Errorf("couldn't connect to config store: %w", err)
	}

	ws, err := workspace.New()
	if err != nil {
		return nil, fmt.Errorf("workspace cannot be created: %w", err)
	}

	p := sessions.NewProvider()
	sess, err := p.Default()
	if err != nil {
		return nil, err
	}

	prompter := prompt.New()
	return &initJobOpts{
		initJobVars: vars,

		fs:          &afero.Afero{Fs: afero.NewOsFs()},
		store:       store,
		ws:          ws,
		appDeployer: cloudformation.New(sess),
		prog:        termprogress.NewSpinner(),
		prompt:      prompter,
		sel:         selector.NewWorkspaceSelect(prompter, store, ws),
	}, nil
}

// Validate returns an error if the flag values passed by the user are invalid.
func (o *initJobOpts) Validate() error {
	if o.jobType != "" {
		if err := validateJobType(o.jobType); err != nil {
			return err
		}
	}
	if o.name != "" {
		if err := validateJobName(o.name); err != nil {
			return err
		}
	}
	if o.dockerfilePath != "" && o.image != "" {
		return fmt.Errorf("--%s and --%s cannot be specified together", dockerFileFlag, imageFlag)
	}
	if o.dockerfilePath != "" {
		if _, err := o.fs.Stat(o.dockerfilePath); err != nil {
			return err
		}
	}
	if o.schedule != "" {
		if err := validateSchedule(o.schedule); err != nil {
			return err
		}
	}
	if o.timeout != "" {
		if err := validateTimeout(o.timeout); err != nil {
			return err
		}
	}
	if o.retries < 0 {
		return errors.New("number of retries must be non-negative")
	}
	return nil
}

// Ask prompts for fields that are required but not passed in.
func (o *initJobOpts) Ask() error {
	if err := o.askJobType(); err != nil {
		return err
	}
	if err := o.askJobName(); err != nil {
		return err
	}
	dfSelected, err := o.askDockerfile()
	if err != nil {
		return err
	}
	if !dfSelected {
		if err := o.askImage(); err != nil {
			return err
		}
	}
	if err := o.askSchedule(); err != nil {
		return err
	}
	return nil
}

// Execute writes the job's manifest file and stores the name in SSM.
func (o *initJobOpts) Execute() error {
	app, err := o.store.GetApplication(o.appName)
	if err != nil {
		return fmt.Errorf("get application %s: %w", o.appName, err)
	}

	manifestPath, err := o.createManifest()
	if err != nil {
		return err
	}
	o.manifestPath = manifestPath

	o.prog.Start(fmt.Sprintf(fmtAddJobToAppStart, o.name))
	if err := o.appDeployer.AddJobToApp(app, o.name); err != nil {
		o.prog.Stop(log.Serrorf(fmtAddJobToAppFailed, o.name))
		return fmt.Errorf("add job %s to application %s: %w", o.name, o.appName, err)
	}
	o.prog.Stop(log.Ssuccessf(fmtAddJobToAppComplete, o.name))

	if err := o.store.CreateJob(&config.Workload{
		App:  o.appName,
		Name: o.name,
		Type: o.jobType,
	}); err != nil {
		return fmt.Errorf("saving job %s: %w", o.name, err)
	}
	return nil
}

func (o *initJobOpts) createManifest() (string, error) {
	manifest, err := o.newJobManifest()
	if err != nil {
		return "", err
	}
	var manifestExists bool
	manifestPath, err := o.ws.WriteJobManifest(manifest, o.name)
	if err != nil {
		e, ok := err.(*workspace.ErrFileExists)
		if !ok {
			return "", err
		}
		manifestExists = true
		manifestPath = e.FileName
	}
	manifestPath, err = relPath(manifestPath)
	if err != nil {
		return "", err
	}

	manifestMsgFmt := "Wrote the manifest for job %s at %s\n"
	if manifestExists {
		manifestMsgFmt = "Manifest file for job %s already exists at %s, skipping writing it.\n"
	}
	log.Successf(manifestMsgFmt, color.HighlightUserInput(o.name), color.HighlightResource(manifestPath))
	log.Infoln(color.Help(fmt.Sprintf("Your manifest contains configurations like your container size and job schedule (%s).", o.schedule)))
	log.Infoln()

	return manifestPath, nil
}

func (o *initJobOpts) newJobManifest() (*manifest.ScheduledJob, error) {
	var dfPath string
	if o.dockerfilePath != "" {
		path, err := relativeDockerfilePath(o.ws, o.dockerfilePath)
		if err != nil {
			return nil, err
		}
		dfPath = path
	}
	return manifest.NewScheduledJob(&manifest.ScheduledJobProps{
		WorkloadProps: &manifest.WorkloadProps{
			Name:       o.name,
			Dockerfile: dfPath,
			Image:      o.image,
		},
		Schedule: o.schedule,
		Timeout:  o.timeout,
		Retries:  o.retries,
	}), nil
}

func (o *initJobOpts) askJobType() error {
	if o.jobType != "" {
		return nil
	}
	// short circuit since there's only one valid job type.
	o.jobType = manifest.ScheduledJobType
	return nil
}

func (o *initJobOpts) askJobName() error {
	if o.name != "" {
		return nil
	}

	name, err := o.prompt.Get(
		fmt.Sprintf(fmtWkldInitNamePrompt, color.Emphasize("name"), color.HighlightUserInput(o.jobType)),
		fmt.Sprintf(fmtWkldInitNameHelpPrompt, job, o.appName),
		validateSvcName,
		prompt.WithFinalMessage("Job name:"),
	)
	if err != nil {
		return fmt.Errorf("get job name: %w", err)
	}
	o.name = name
	return nil
}

func (o *initJobOpts) askImage() error {
	if o.image != "" {
		return nil
	}
	image, err := o.prompt.Get(wkldInitImagePrompt, wkldInitImagePromptHelp, nil,
		prompt.WithFinalMessage("Image:"))
	if err != nil {
		return fmt.Errorf("get image location: %w", err)
	}
	o.image = image
	return nil
}

// isDfSelected indicates if any Dockerfile is in use.
func (o *initJobOpts) askDockerfile() (isDfSelected bool, err error) {
	if o.dockerfilePath != "" || o.image != "" {
		return true, nil
	}
	df, err := o.sel.Dockerfile(
		fmt.Sprintf(fmtWkldInitDockerfilePrompt, color.HighlightUserInput(o.name)),
		fmt.Sprintf(fmtWkldInitDockerfilePathPrompt, color.HighlightUserInput(o.name)),
		wkldInitDockerfileHelpPrompt,
		wkldInitDockerfilePathHelpPrompt,
		func(v interface{}) error {
			return validatePath(afero.NewOsFs(), v)
		},
	)
	if err != nil {
		return false, fmt.Errorf("select Dockerfile: %w", err)
	}
	if df == selector.DockerfilePromptUseImage {
		return false, nil
	}
	o.dockerfilePath = df
	return true, nil
}

func (o *initJobOpts) askSchedule() error {
	if o.schedule != "" {
		return nil
	}
	schedule, err := o.sel.Schedule(
		jobInitSchedulePrompt,
		jobInitScheduleHelp,
		validateSchedule,
		validateRate,
	)
	if err != nil {
		return fmt.Errorf("get schedule: %w", err)
	}

	o.schedule = schedule
	return nil
}

// RecommendedActions returns follow-up actions the user can take after successfully executing the command.
func (o *initJobOpts) RecommendedActions() []string {
	return []string{
		fmt.Sprintf("Update your manifest %s to change the defaults.", color.HighlightResource(o.manifestPath)),
		fmt.Sprintf("Run %s to deploy your job to a %s environment.",
			color.HighlightCode(fmt.Sprintf("copilot job deploy --name %s --env %s", o.name, defaultEnvironmentName)),
			defaultEnvironmentName),
	}
}

// buildJobInitCmd builds the command for creating a new job.
func buildJobInitCmd() *cobra.Command {
	vars := initJobVars{}
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Creates a new scheduled job in an application.",
		Example: `
  Create a "reaper" scheduled task to run once per day.
  /code $ copilot job init --name reaper --dockerfile ./frontend/Dockerfile --schedule "every 2 hours"

  Create a "report-generator" scheduled task with retries.
  /code $ copilot job init --name report-generator --schedule "@monthly" --retries 3 --timeout 900s`,
		RunE: runCmdE(func(cmd *cobra.Command, args []string) error {
			opts, err := newInitJobOpts(vars)
			if err != nil {
				return err
			}
			if err := opts.Validate(); err != nil { // validate flags
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
	cmd.Flags().StringVarP(&vars.jobType, jobTypeFlag, jobTypeFlagShort, "", jobTypeFlagDescription)
	cmd.Flags().StringVarP(&vars.dockerfilePath, dockerFileFlag, dockerFileFlagShort, "", dockerFileFlagDescription)
	cmd.Flags().StringVarP(&vars.schedule, scheduleFlag, scheduleFlagShort, "", scheduleFlagDescription)
	cmd.Flags().StringVar(&vars.timeout, timeoutFlag, "", timeoutFlagDescription)
	cmd.Flags().IntVar(&vars.retries, retriesFlag, 0, retriesFlagDescription)
	cmd.Flags().StringVarP(&vars.image, imageFlag, imageFlagShort, "", imageFlagDescription)

	cmd.Annotations = map[string]string{
		"group": group.Develop,
	}

	return cmd
}
