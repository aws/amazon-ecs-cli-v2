// Copyright 2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"errors"
	"fmt"

	"github.com/aws/amazon-ecs-cli-v2/internal/pkg/archer"
	"github.com/aws/amazon-ecs-cli-v2/internal/pkg/manifest"
	"github.com/aws/amazon-ecs-cli-v2/internal/pkg/store"
	"github.com/aws/amazon-ecs-cli-v2/internal/pkg/store/secretsmanager"
	"github.com/aws/amazon-ecs-cli-v2/internal/pkg/term/color"
	"github.com/aws/amazon-ecs-cli-v2/internal/pkg/term/log"
	"github.com/aws/amazon-ecs-cli-v2/internal/pkg/workspace"

	"github.com/spf13/cobra"
)

const (
	pipelineAddEnvPrompt          = "Would you like to add an environment to your pipeline?"
	pipelineSelectEnvPrompt       = "Which environment would you like to add to your pipeline?"
	pipelineEnterGitHubRepoPrompt = "What is your application's GitHub repository?" // TODO allow just <user>/<repo>?
	pipelineDefaultFilename       = "pipeline.yml"
)

var errNoEnvsInProject = errors.New("There were no more environments found that can be added to your pipeline. Please run `archer env init` to create a new environment.")

// InitPipelineOpts holds the configuration needed to create a new pipeilne
type InitPipelineOpts struct {
	// Fields with matching flags.
	Environments      []string
	GitHubRepo        string
	GitHubAccessToken string
	EnableCD          bool
	Deploy            bool
	PipelineFilename  string
	// TODO add git branch
	// TODO add pipeline file (to write to different file than pipeline.yml?)

	// Interfaces to interact with dependencies.
	manifestWriter archer.ManifestIO
	secretsmanager archer.SecretsManager

	// Outputs stored on successful actions.
	manifestPath string
	secretArn    string
	secretName   string

	// Caches environments
	projectEnvs []string

	*GlobalOpts
}

func NewInitPipelineOpts() *InitPipelineOpts {
	return &InitPipelineOpts{
		GlobalOpts: NewGlobalOpts(),
	}
}

// Ask prompts for fields that are required but not passed in.
func (opts *InitPipelineOpts) Ask() error {
	if len(opts.Environments) == 0 {
		if err := opts.selectEnvironments(true); err != nil {
			return err
		}
	}

	if opts.GitHubRepo == "" {
		if err := opts.selectGitHubRepo(); err != nil {
			return err
		}
	}

	if opts.GitHubAccessToken == "" {
		if err := opts.getGitHubAccessToken(); err != nil {
			return err
		}
	}

	// if err := opts.askEnableCD(); err != nil {
	// 	return err
	// }

	// TODO ask this after pipeline.yml is written
	// if err := opts.askDeploy(); err != nil {
	// 	return err
	// }

	return nil
}

// Validate returns an error if the flag values passed by the user are invalid.
func (opts *InitPipelineOpts) Validate() error {
	// TODO
	if opts.ProjectName() == "" {
		return errNoProjectInWorkspace
	}

	if len(opts.projectEnvs) == 0 {
		return errNoEnvsInProject
	}

	return nil
}

// Execute writes the pipline manifest file.
func (opts *InitPipelineOpts) Execute() error {
	secretName := opts.createSecretName()
	secretArn, err := opts.secretsmanager.CreateSecret(secretName, opts.GitHubAccessToken)
	if err != nil {
		return err
	}
	opts.secretArn = secretArn
	opts.secretName = secretName
	log.Successf("Created secret: %s for GitHub repo: %s'\n", color.HighlightResource(secretName), color.HighlightResource(opts.GitHubRepo))

	// write pipeline.yml file, populate with:
	//   - github repo as source
	//   - stage names (environments)
	//   - enable/disable transition to prod envs

	manifestPath, err := opts.createPipelineManifest()
	if err != nil {
		return err
	}
	opts.manifestPath = manifestPath

	// TODO create buildspec file

	log.Infoln()
	log.Successf("Wrote the pipeline for %s app at '%s'\n", color.HighlightUserInput(opts.GitHubRepo), color.HighlightResource(opts.manifestPath))
	log.Infoln("Your pipeline manifest contains configurations for your CodePipeline resources, such as your pipeline stages and build steps.")
	log.Infoln()

	// TODO deploy manifest file

	return nil
}

func (opts *InitPipelineOpts) createSecretName() string {
	repoName := opts.getRepoName()
	return fmt.Sprintf("github-token-%s-%s", opts.projectName, repoName)
}

func (opts *InitPipelineOpts) createPipelineName() string {
	repoName := opts.getRepoName()
	return fmt.Sprintf("pipeline-%s-%s", opts.projectName, repoName)
}

func (opts *InitPipelineOpts) getRepoName() string {
	match := githubRepoExp.FindStringSubmatch(opts.GitHubRepo)
	if len(match) == 0 {
		return ""
	}

	matches := make(map[string]string)
	for i, name := range githubRepoExp.SubexpNames() {
		if i != 0 && name != "" {
			matches[name] = match[i]
		}
	}

	return matches["repo"]
}

func (opts *InitPipelineOpts) createPipelineProvider() (manifest.Provider, error) {
	config := &manifest.GitHubProperties{
		OwnerAndRepository:    opts.GitHubRepo,
		Branch:                "master", // todo - fix
		GithubSecretIdKeyName: opts.secretName,
	}

	return manifest.NewProvider(config)
}

func (opts *InitPipelineOpts) createPipelineManifest() (string, error) {
	// TODO change this to flag
	pipelineName := opts.createPipelineName()
	provider, err := opts.createPipelineProvider()
	if err != nil {
		return "", fmt.Errorf("could not create pipeline: %w", err)
	}

	manifest, err := manifest.CreatePipeline(pipelineName, provider, opts.Environments)
	if err != nil {
		return "", fmt.Errorf("generate a manifest: %w", err)
	}

	manifestBytes, err := manifest.Marshal()
	if err != nil {
		return "", fmt.Errorf("marshal manifest: %w", err)
	}
	manifestPath, err := opts.manifestWriter.WriteManifest(manifestBytes, pipelineDefaultFilename)
	if err != nil {
		return "", fmt.Errorf("write manifest for app %s: %w", pipelineDefaultFilename, err)
	}

	return manifestPath, nil
}

func (opts *InitPipelineOpts) selectEnvironments(addMore bool) error {
	if addMore == false {
		return nil
	}

	addEnv, err := opts.prompt.Confirm(
		pipelineAddEnvPrompt,
		"Adds an environment that corresponds to a deployment stage in your pipeline. Environments are added sequentially.",
	)

	if err != nil {
		return fmt.Errorf("failed to confirm adding an environment: %w", err)
	}

	var selectMoreEnvs bool
	if addEnv {
		selectMore, err := opts.selectEnvironment()
		if err != nil {
			return err
		}
		selectMoreEnvs = selectMore
	}

	return opts.selectEnvironments(selectMoreEnvs)
}

func (opts *InitPipelineOpts) listAvailableEnvironments() []string {
	envs := []string{}
	for _, env := range opts.projectEnvs {
		// Check if environment has already been added to pipeline
		if opts.envCanBeAdded(env) {
			envs = append(envs, env)
		}
	}

	return envs
}

func (opts *InitPipelineOpts) envCanBeAdded(selectedEnv string) bool {
	for _, env := range opts.Environments {
		if selectedEnv == env {
			return false
		}
	}

	return true
}

func (opts *InitPipelineOpts) selectEnvironment() (bool, error) {
	selectMoreEnvs := false

	envs := opts.listAvailableEnvironments()

	if len(envs) == 0 && len(opts.Environments) != 0 {
		log.Infoln("There are no more environments to add.")
		return selectMoreEnvs, nil
	}

	env, err := opts.prompt.SelectOne(
		pipelineSelectEnvPrompt,
		"Environment to be added as the next stage in your pipeline.",
		envs,
	)

	if err != nil {
		return selectMoreEnvs, fmt.Errorf("failed to add environment: %w", err)
	}

	opts.Environments = append(opts.Environments, env)
	selectMoreEnvs = true

	return selectMoreEnvs, nil
}

// TODO: Nice-to-have: have an opts.listRemoteRepos() method that execs out to `git remote -v` and parse repo name to offer select menu
func (opts *InitPipelineOpts) selectGitHubRepo() error {
	repo, err := opts.prompt.Get(
		pipelineEnterGitHubRepoPrompt,
		fmt.Sprintf(`The GitHub repository linked to your workspace. Pushing to this repository will trigger your pipeline build stage. Please enter full repository URL, e.g. https://github.com/myCompany/myRepo`),
		validateGitHubRepo,
	)

	if err != nil {
		return fmt.Errorf("failed to get GitHub repository: %w", err)
	}

	opts.GitHubRepo = repo

	return nil
}

func (opts *InitPipelineOpts) getGitHubAccessToken() error {
	token, err := opts.prompt.GetSecret(
		fmt.Sprintf("Please enter your GitHub Personal Access Token for your repository: %s", opts.GitHubRepo),
		fmt.Sprintf(`The personal access token for the GitHub repository linked to your workspace. For more information on how to create a personal access token, please refer to: https://help.github.com/en/enterprise/2.17/user/authenticating-to-github/creating-a-personal-access-token-for-the-command-line.`),
	)

	if err != nil {
		return fmt.Errorf("failed to get GitHub access token: %w", err)
	}

	opts.GitHubAccessToken = token

	return nil
}

func (opts *InitPipelineOpts) askEnableCD() error {
	enable, err := opts.prompt.Confirm(
		"Would you like to automatically enable deploying to production?",
		"Enables the transition to your production environment automatically through your pipeline.",
	)

	if err != nil {
		return fmt.Errorf("failed to confirm enabling CD: %w", err)
	}

	opts.EnableCD = enable

	return nil
}

func (opts *InitPipelineOpts) askDeploy() error {
	deploy, err := opts.prompt.Confirm(
		"Would you like to deploy your pipeline?",
		"Deploys your pipeline through CloudFormation.",
	)

	if err != nil {
		return fmt.Errorf("failed to confirm deploying pipeline: %w", err)
	}

	opts.Deploy = deploy

	return nil
}

func (opts *InitPipelineOpts) getEnvNames() ([]string, error) {
	store, err := store.New()
	if err != nil {
		return nil, fmt.Errorf("couldn't connect to environment datastore: %w", err)
	}

	envs, err := store.ListEnvironments(opts.ProjectName())
	if err != nil {
		return nil, fmt.Errorf("could not list environments for project %s: %w", opts.ProjectName(), err)
	}

	if len(envs) == 0 {
		return nil, errNoEnvsInProject
	}

	envNames := []string{}
	for _, env := range envs {
		envNames = append(envNames, env.Name)
	}

	return envNames, nil
}

// BuildPipelineInitCmd build the command for creating a new pipeline.
func BuildPipelineInitCmd() *cobra.Command {
	opts := NewInitPipelineOpts()
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Creates a pipeline for applications in your workspace.",
		Long:  `Creates a pipeline for the applications in your workspace, using the environments associated with the applications.`,
		Example: `
  Create a pipeline for the applications in your workspace:
  /code $ archer pipeline init \
    --github-repo "gitHubUserName/myFrontendApp" \
    --github-access-token file://myGitHubToken \
    --environments "stage,prod" \
    --deploy`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			projectEnvs, err := opts.getEnvNames()
			if err != nil {
				return fmt.Errorf("couldn't get environments: %w", err)
			}
			opts.projectEnvs = projectEnvs

			ws, err := workspace.New()
			if err != nil {
				return fmt.Errorf("workspace cannot be created: %w", err)
			}
			opts.manifestWriter = ws

			secretsmanager, err := secretsmanager.NewStore()
			if err != nil {
				return fmt.Errorf("couldn't create secrets manager: %w", err)
			}
			opts.secretsmanager = secretsmanager

			return opts.Validate()
		},

		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.Ask(); err != nil {
				return err
			}
			if err := opts.Validate(); err != nil { // validate flags
				return err
			}
			return opts.Execute()
		},
		// TODO: recommend editing pipeline.yml and deploying pipeline
		// PostRunE: func(cmd *cobra.Command, args []string) error {
		// 	log.Infoln("Recommended follow-up actions:")
		// 	for _, followup := range opts.RecommendedActions() {
		// 		log.Infof("- %s\n", followup)
		// 	}
		// 	return nil
		// },
	}
	cmd.Flags().StringVarP(&opts.GitHubRepo, githubRepoFlag, githubRepoFlagShort, "", githubRepoFlagDescription)
	cmd.Flags().StringVarP(&opts.GitHubAccessToken, githubAccessTokenFlag, githubAccessTokenFlagShort, "", githubAccessTokenFlagDescription)
	cmd.Flags().BoolVar(&opts.Deploy, deployFlag, false, deployPipelineFlagDescription)
	cmd.Flags().BoolVar(&opts.EnableCD, enableCDFlag, false, enableCDFlagDescription)
	cmd.Flags().StringSliceVarP(&opts.Environments, envsFlag, envsFlagShort, []string{}, pipelineEnvsFlagDescription)

	return cmd
}
