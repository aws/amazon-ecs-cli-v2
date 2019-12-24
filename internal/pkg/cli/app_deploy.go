// Copyright 2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/aws/amazon-ecs-cli-v2/internal/pkg/archer"
	"github.com/aws/amazon-ecs-cli-v2/internal/pkg/aws/ecr"
	"github.com/aws/amazon-ecs-cli-v2/internal/pkg/aws/session"
	"github.com/aws/amazon-ecs-cli-v2/internal/pkg/build/docker"
	"github.com/aws/amazon-ecs-cli-v2/internal/pkg/deploy/cloudformation"
	"github.com/aws/amazon-ecs-cli-v2/internal/pkg/deploy/cloudformation/stack"
	"github.com/aws/amazon-ecs-cli-v2/internal/pkg/describe"
	"github.com/aws/amazon-ecs-cli-v2/internal/pkg/manifest"
	"github.com/aws/amazon-ecs-cli-v2/internal/pkg/store"
	"github.com/aws/amazon-ecs-cli-v2/internal/pkg/term/color"
	"github.com/aws/amazon-ecs-cli-v2/internal/pkg/term/command"
	"github.com/aws/amazon-ecs-cli-v2/internal/pkg/term/log"
	termprogress "github.com/aws/amazon-ecs-cli-v2/internal/pkg/term/progress"
	"github.com/aws/amazon-ecs-cli-v2/internal/pkg/workspace"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

const (
	inputImageTagPrompt = "Input an image tag value:"
)

var (
	errNoLocalManifestsFound = errors.New("no manifest files found")
)

type appDeployOpts struct {
	*GlobalOpts
	app      string
	env      string
	imageTag string

	projectService     projectService
	workspaceService   archer.Workspace
	ecrService         ecrService
	dockerService      dockerService
	runner             runner
	appPackageCfClient projectResourcesGetter
	appDeployCfClient  cloudformation.CloudFormation
	sessProvider       sessionProvider

	spinner progress

	localProjectAppNames []string
	projectEnvironments  []*archer.Environment

	targetEnvironment *archer.Environment
}

func (opts appDeployOpts) String() string {
	return fmt.Sprintf("project: %s, app: %s, env: %s, tag: %s", opts.ProjectName(), opts.app, opts.env, opts.imageTag)
}

type projectService interface {
	archer.ProjectStore
	archer.EnvironmentStore
	archer.ApplicationStore
}

type ecrService interface {
	GetRepository(name string) (string, error)
	GetECRAuth() (ecr.Auth, error)
}

type dockerService interface {
	Build(uri, tag, path string) error
	Login(uri, username, password string) error
	Push(uri, tag string) error
}

type runner interface {
	Run(name string, args []string, options ...command.Option) error
}

func (opts *appDeployOpts) init() error {
	projectService, err := store.New()
	if err != nil {
		return fmt.Errorf("create project service: %w", err)
	}
	opts.projectService = projectService

	workspaceService, err := workspace.New()
	if err != nil {
		return fmt.Errorf("intialize workspace service: %w", err)
	}
	opts.workspaceService = workspaceService
	return nil
}

func (opts *appDeployOpts) sourceInputs() error {
	if opts.ProjectName() == "" {
		return errNoProjectInWorkspace
	}

	if err := opts.sourceProjectData(); err != nil {
		return err
	}

	if err := opts.sourceAppName(); err != nil {
		return err
	}

	if err := opts.sourceTargetEnv(); err != nil {
		return err
	}

	if err := opts.configureClients(); err != nil {
		return err
	}

	if err := opts.sourceImageTag(); err != nil {
		return err
	}

	return nil
}

func (opts *appDeployOpts) sourceProjectData() error {
	if err := opts.sourceProjectApplications(); err != nil {
		return err
	}

	if err := opts.sourceProjectEnvironments(); err != nil {
		return err
	}

	return nil
}

func (opts *appDeployOpts) sourceProjectApplications() error {
	apps, err := opts.workspaceService.Apps()

	if err != nil {
		return fmt.Errorf("get apps: %w", err)
	}

	if len(apps) == 0 {
		// TODO: recommend follow up command - app init?
		return errors.New("no applications found")
	}

	names := make([]string, 0, len(apps))
	for _, app := range apps {
		names = append(names, app.AppName())
	}

	opts.localProjectAppNames = names

	return nil
}

func (opts *appDeployOpts) sourceProjectEnvironments() error {
	envs, err := opts.projectService.ListEnvironments(opts.ProjectName())

	if err != nil {
		return fmt.Errorf("get environments: %w", err)
	}

	if len(envs) == 0 {
		// TODO: recommend follow up command - env init?
		log.Infof("couldn't find any environments associated with project %s, try initializing one: %s\n",
			color.HighlightUserInput(opts.ProjectName()),
			color.HighlightCode("ecs-preview env init"))

		return errors.New("no environments found")
	}

	opts.projectEnvironments = envs

	return nil
}

func (opts *appDeployOpts) sourceAppName() error {
	if opts.app == "" {
		if len(opts.localProjectAppNames) == 1 {
			opts.app = opts.localProjectAppNames[0]

			// NOTE: defaulting the app name, tell the user
			log.Infof("Only found one app, defaulting to: %s\n", color.HighlightUserInput(opts.app))

			return nil
		}

		selectedAppName, err := opts.prompt.SelectOne("Select an application", "", opts.localProjectAppNames)

		if err != nil {
			return fmt.Errorf("select app name: %w", err)
		}

		opts.app = selectedAppName
	}

	for _, appName := range opts.localProjectAppNames {
		if opts.app == appName {
			return nil
		}
	}

	return fmt.Errorf("invalid app name: %s", opts.app)
}

func (opts *appDeployOpts) sourceTargetEnv() error {
	if opts.env == "" {
		if len(opts.projectEnvironments) == 1 {
			opts.targetEnvironment = opts.projectEnvironments[0]

			// NOTE: defaulting the env name, tell the user.
			log.Infof("Only found one environment, defaulting to: %s\n", color.HighlightUserInput(opts.targetEnvironment.Name))

			return nil
		}

		var envNames []string
		for _, env := range opts.projectEnvironments {
			envNames = append(envNames, env.Name)
		}

		selectedEnvName, err := opts.prompt.SelectOne("Select an environment", "", envNames)

		if err != nil {
			return fmt.Errorf("select env name: %w", err)
		}

		opts.env = selectedEnvName
	}

	for _, env := range opts.projectEnvironments {
		if opts.env == env.Name {
			opts.targetEnvironment = env

			return nil
		}
	}

	return fmt.Errorf("invalid env name: %s", opts.env)
}

func (opts *appDeployOpts) configureClients() error {
	defaultSessEnvRegion, err := opts.sessProvider.DefaultWithRegion(opts.targetEnvironment.Region)
	if err != nil {
		return fmt.Errorf("create ECR session with region %s: %w", opts.targetEnvironment.Region, err)
	}

	envSession, err := opts.sessProvider.FromRole(opts.targetEnvironment.ManagerRoleARN, opts.targetEnvironment.Region)
	if err != nil {
		return fmt.Errorf("assuming environment manager role: %w", err)
	}

	// ECR client against tools account profile AND target environment region
	opts.ecrService = ecr.New(defaultSessEnvRegion)

	// app deploy CF client against env account profile AND target environment region
	opts.appDeployCfClient = cloudformation.New(envSession)

	// app package CF client against tools account
	appPackageCfSess, err := opts.sessProvider.Default()
	if err != nil {
		return fmt.Errorf("create app package CF session: %w", err)
	}
	opts.appPackageCfClient = cloudformation.New(appPackageCfSess)

	return nil
}

func (opts *appDeployOpts) sourceImageTag() error {
	if opts.imageTag != "" {
		return nil
	}

	var buf bytes.Buffer
	err := opts.runner.Run("git", []string{"describe", "--always"}, command.Stdout(&buf))

	if err == nil {
		// NOTE: `git describe` output bytes includes a `\n` character, so we trim it out.
		opts.imageTag = strings.TrimSpace(buf.String())

		return nil
	}

	log.Warningln("Failed to default tag, are you in a git repository?")

	tag, err := opts.prompt.Get(inputImageTagPrompt, "", nil /*no validation*/)
	if err != nil {
		return fmt.Errorf("prompt for image tag: %w", err)
	}

	opts.imageTag = tag

	return nil
}

func (opts appDeployOpts) deployApp() error {
	repoName := fmt.Sprintf("%s/%s", opts.projectName, opts.app)

	uri, err := opts.ecrService.GetRepository(repoName)
	if err != nil {
		return fmt.Errorf("get ECR repository URI: %w", err)
	}

	appDockerfilePath, err := opts.getAppDockerfilePath()
	if err != nil {
		return err
	}

	if err := opts.dockerService.Build(uri, opts.imageTag, appDockerfilePath); err != nil {
		return fmt.Errorf("build Dockerfile at %s with tag %s: %w", appDockerfilePath, opts.imageTag, err)
	}

	auth, err := opts.ecrService.GetECRAuth()

	if err != nil {
		return fmt.Errorf("get ECR auth data: %w", err)
	}

	opts.dockerService.Login(uri, auth.Username, auth.Password)

	if err != nil {
		return err
	}

	if err = opts.dockerService.Push(uri, opts.imageTag); err != nil {
		return err
	}

	template, err := opts.getAppDeployTemplate()

	id, err := uuid.NewRandom()
	if err != nil {
		return fmt.Errorf("failed to generate random id for changeSet: %w", err)
	}
	stackName := stack.NameForApp(opts.ProjectName(), opts.targetEnvironment.Name, opts.app)
	changeSetName := fmt.Sprintf("%s-%s", stackName, id)

	opts.spinner.Start(
		fmt.Sprintf("Deploying %s to %s.",
			fmt.Sprintf("%s:%s", color.HighlightUserInput(opts.app), color.HighlightUserInput(opts.imageTag)),
			color.HighlightUserInput(opts.targetEnvironment.Name)))

	// TODO Use the Tags() method defined in deploy/cloudformation/stack/lb_fargate_app.go
	tags := map[string]string{
		stack.ProjectTagKey: opts.ProjectName(),
		stack.EnvTagKey:     opts.targetEnvironment.Name,
		stack.AppTagKey:     opts.app,
	}
	if err := opts.applyAppDeployTemplate(template, stackName, changeSetName, opts.targetEnvironment.ExecutionRoleARN, tags); err != nil {
		opts.spinner.Stop("Error!")
		return err
	}
	opts.spinner.Stop("")

	identifier, err := describe.NewAppIdentifier(opts.ProjectName(), opts.app)
	if err != nil {
		return fmt.Errorf("create identifier for application %s in project %s: %w", opts.app, opts.ProjectName(), err)
	}
	uri, err = identifier.URI(opts.targetEnvironment.Name)
	if err != nil {
		return fmt.Errorf("cannot retrieve the URI from environment %s: %w", opts.env, err)
	}
	log.Successf("Deployed %s, you can access it at %s\n", color.HighlightUserInput(opts.app), uri)

	return nil
}

func (opts appDeployOpts) getAppDeployTemplate() (string, error) {
	buffer := &bytes.Buffer{}

	appPackage := PackageAppOpts{
		AppName:      opts.app,
		EnvName:      opts.targetEnvironment.Name,
		Tag:          opts.imageTag,
		stackWriter:  buffer,
		paramsWriter: ioutil.Discard,
		store:        opts.projectService,
		describer:    opts.appPackageCfClient,
		ws:           opts.workspaceService,
		GlobalOpts:   opts.GlobalOpts,
	}

	if err := appPackage.Execute(); err != nil {
		return "", fmt.Errorf("package application: %w", err)
	}

	return buffer.String(), nil
}

func (opts appDeployOpts) applyAppDeployTemplate(template, stackName, changeSetName, cfExecutionRole string, tags map[string]string) error {
	if err := opts.appDeployCfClient.DeployApp(template, stackName, changeSetName, cfExecutionRole, tags); err != nil {
		return fmt.Errorf("deploy application: %w", err)
	}

	return nil
}

func (opts appDeployOpts) getAppDockerfilePath() (string, error) {
	manifestFileNames, err := opts.workspaceService.ListManifestFiles()
	if err != nil {
		return "", fmt.Errorf("list local manifest files: %w", err)
	}
	if len(manifestFileNames) == 0 {
		return "", errNoLocalManifestsFound
	}

	var targetManifestFile string
	for _, f := range manifestFileNames {
		if strings.Contains(f, opts.app) {
			targetManifestFile = f
			break
		}
	}
	if targetManifestFile == "" {
		return "", fmt.Errorf("couldn't find local manifest %s", opts.app)
	}

	manifestBytes, err := opts.workspaceService.ReadFile(targetManifestFile)
	if err != nil {
		return "", fmt.Errorf("read manifest file %s: %w", targetManifestFile, err)
	}

	mf, err := manifest.UnmarshalApp(manifestBytes)
	if err != nil {
		return "", fmt.Errorf("unmarshal app manifest: %w", err)
	}

	return strings.TrimSuffix(mf.DockerfilePath(), "/Dockerfile"), nil
}

// BuildAppDeployCmd builds the `app deploy` subcommand.
func BuildAppDeployCmd() *cobra.Command {
	input := &appDeployOpts{
		GlobalOpts:    NewGlobalOpts(),
		spinner:       termprogress.NewSpinner(),
		dockerService: docker.New(),
		runner:        command.New(),
		sessProvider:  session.NewProvider(),
	}

	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy an application to an environment.",
		Long:  `Deploy an application to an environment.`,
		Example: `
  Deploy an application named "frontend" to a "test" environment.
  /code $ ecs-preview app deploy --name frontend --env test`,
		PreRunE: runCmdE(func(cmd *cobra.Command, args []string) error {
			if err := input.init(); err != nil {
				return err
			}
			if err := input.sourceInputs(); err != nil {
				return err
			}
			return nil
		}),
		RunE: runCmdE(func(cmd *cobra.Command, args []string) error {
			if err := input.deployApp(); err != nil {
				return err
			}
			return nil
		}),
		// PostRunE: func(cmd *cobra.Command, args []string) error {
		// TODO: recommended actions?
		// },
	}

	cmd.Flags().StringVarP(&input.app, nameFlag, nameFlagShort, "", appFlagDescription)
	cmd.Flags().StringVarP(&input.env, envFlag, envFlagShort, "", envFlagDescription)
	cmd.Flags().StringVar(&input.imageTag, imageTagFlag, "", imageTagFlagDescription)

	return cmd
}
