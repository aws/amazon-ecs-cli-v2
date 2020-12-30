// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/aws/copilot-cli/internal/pkg/term/color"
	"github.com/aws/copilot-cli/internal/pkg/term/log"
	"github.com/aws/copilot-cli/internal/pkg/version"
	"github.com/spf13/cobra"

	"github.com/dustin/go-humanize"

	"github.com/aws/copilot-cli/internal/pkg/term/selector"

	"github.com/aws/copilot-cli/internal/pkg/aws/secretsmanager"
	"github.com/aws/copilot-cli/internal/pkg/aws/sessions"
	"github.com/aws/copilot-cli/internal/pkg/config"
	"github.com/aws/copilot-cli/internal/pkg/deploy/cloudformation"
	"github.com/aws/copilot-cli/internal/pkg/manifest"
	"github.com/aws/copilot-cli/internal/pkg/template"
	"github.com/aws/copilot-cli/internal/pkg/term/command"
	"github.com/aws/copilot-cli/internal/pkg/term/prompt"
	"github.com/aws/copilot-cli/internal/pkg/workspace"
	"github.com/spf13/afero"
)

const (
	pipelineSelectEnvPrompt     = "Which environment would you like to add to your pipeline?"
	pipelineSelectEnvHelpPrompt = "Adds an environment that corresponds to a deployment stage in your pipeline. Environments are added sequentially."

	pipelineSelectURLPrompt     = "Which repository would you like to use for your service?"
	pipelineSelectURLHelpPrompt = `The repository linked to your workspace.
Pushing to this repository will trigger your pipeline build stage.
Please enter full repository URL, e.g. "https://github.com/myCompany/myRepo", or the owner/rep, e.g. "myCompany/myRepo"`
)

const (
	buildspecTemplatePath = "cicd/buildspec.yml"
	githubURL             = "github.com"
	codeCommit            = "codecommit"
	awsURL                = "aws.amazon.com"
	ghProviderName        = "GitHub"
	ccProviderName        = "CodeCommit"
	defaultGHBranch       = "main"
	defaultCCBranch       = "master"
)

const (
	fmtSecretName         = "github-token-%s-%s"
	fmtGHPipelineName     = "pipeline-%s-%s-%s"
	fmtCCPipelineName     = "pipeline-%s-%s"
	fmtGHPipelineProvider = "https://%s/%s/%s"
	fmtCCPipelineProvider = "https://%s.console.%s/codesuite/codecommit/repositories/%s"
)

var (
	// Filled in via the -ldflags flag at compile time to support pipeline buildspec CLI pulling.
	binaryS3BucketPath string
)

type initPipelineVars struct {
	appName           string
	environments      []string
	repoURL           string
	repoBranch        string
	githubAccessToken string
}

type initPipelineOpts struct {
	initPipelineVars
	// Interfaces to interact with dependencies.
	workspace      wsPipelineWriter
	secretsmanager secretsManager
	parser         template.Parser
	runner         runner
	cfnClient      appResourcesGetter
	store          store
	prompt         prompter
	sel            pipelineSelector

	// Outputs stored on successful actions.
	secret      string
	provider    string
	repoName    string
	githubOwner string
	ccRegion    string

	// Caches variables
	fs         *afero.Afero
	buffer     bytes.Buffer
	envConfigs []*config.Environment
}

type artifactBucket struct {
	BucketName   string
	Region       string
	Environments []string
}

func newInitPipelineOpts(vars initPipelineVars) (*initPipelineOpts, error) {
	ws, err := workspace.New()
	if err != nil {
		return nil, fmt.Errorf("new workspace client: %w", err)
	}

	secretsmanager, err := secretsmanager.New()
	if err != nil {
		return nil, fmt.Errorf("new secretsmanager client: %w", err)
	}

	p := sessions.NewProvider()
	defaultSession, err := p.Default()
	if err != nil {
		return nil, err
	}

	ssmStore, err := config.NewStore()
	if err != nil {
		return nil, fmt.Errorf("new config store client: %w", err)
	}

	prompter := prompt.New()

	return &initPipelineOpts{
		initPipelineVars: vars,
		workspace:        ws,
		secretsmanager:   secretsmanager,
		parser:           template.New(),
		cfnClient:        cloudformation.New(defaultSession),
		store:            ssmStore,
		prompt:           prompter,
		sel:              selector.NewSelect(prompter, ssmStore),
		runner:           command.New(),
		fs:               &afero.Afero{Fs: afero.NewOsFs()},
	}, nil
}

// Validate returns an error if the flag values passed by the user are invalid.
func (o *initPipelineOpts) Validate() error {
	if o.appName == "" {
		return errNoAppInWorkspace
	}
	if o.appName != "" {
		if _, err := o.store.GetApplication(o.appName); err != nil {
			return err
		}
	}

	if o.repoURL != "" {
		if err := o.validateURL(o.repoURL); err != nil {
			return err
		}
	}

	if o.environments != nil {
		for _, env := range o.environments {
			_, err := o.store.GetEnvironment(o.appName, env)
			if err != nil {
				return err
			}
		}
		return nil
	}
	return nil
}

// Ask prompts for fields that are required but not passed in.
func (o *initPipelineOpts) Ask() error {
	if err := o.askEnvs(); err != nil {
		return err
	}
	if err := o.askRepository(); err != nil {
		return err
	}
	return nil
}

// Execute writes the pipeline manifest file.
func (o *initPipelineOpts) Execute() error {
	if o.provider == ghProviderName {
		secretName := o.secretName()
		_, err := o.secretsmanager.CreateSecret(secretName, o.githubAccessToken)

		if err != nil {
			var existsErr *secretsmanager.ErrSecretAlreadyExists
			if !errors.As(err, &existsErr) {
				return err
			}
			log.Successf("Secret already exists for %s! Do nothing.\n", color.HighlightUserInput(o.repoName))
		} else {
			log.Successf("Created the secret %s for pipeline source stage!\n", color.HighlightUserInput(secretName))
		}
		o.secret = secretName
	}

	// write pipeline.yml file, populate with:
	//   - git repo as source
	//   - stage names (environments)
	//   - enable/disable transition to prod envs
	var err error
	err = o.createPipelineManifest()
	if err != nil {
		return err
	}

	err = o.createBuildspec()
	if err != nil {
		return err
	}
	return nil
}

// RecommendedActions returns follow-up actions the user can take after successfully executing the command.
func (o *initPipelineOpts) RecommendedActions() []string {
	return []string{
		"Commit and push the generated buildspec and manifest file.",
		fmt.Sprintf("Update the %s phase of your buildspec to unit test your services before pushing the images.", color.HighlightResource("build")),
		"Update your pipeline manifest to add additional stages.",
		fmt.Sprintf("Run %s to deploy your pipeline for the repository.", color.HighlightCode("copilot pipeline update")),
	}
}

func (o *initPipelineOpts) validateURL(url string) error {
	// Note: no longer calling `validateDomainName` because if users use git-remote-codecommit
	// (the HTTPS (GRC) protocol) to connect to CodeCommit, the url does not have any periods.
	if !strings.Contains(url, githubURL) && !strings.Contains(url, codeCommit) {
		return errors.New("Copilot currently accepts only URLs to GitHub and CodeCommit repository sources")
	}
	return nil
}

func (o *initPipelineOpts) askEnvs() error {
	if len(o.environments) == 0 {
		envs, err := o.sel.Environments(pipelineSelectEnvPrompt, pipelineSelectEnvHelpPrompt, o.appName, func(order int) prompt.Option {
			return prompt.WithFinalMessage(fmt.Sprintf("%s stage:", humanize.Ordinal(order)))
		})
		if err != nil {
			return fmt.Errorf("select environments: %w", err)
		}
		o.environments = envs
	}

	var envConfigs []*config.Environment
	for _, environment := range o.environments {
		envConfig, err := o.store.GetEnvironment(o.appName, environment)
		if err != nil {
			return fmt.Errorf("get config of environment: %w", err)
		}
		envConfigs = append(envConfigs, envConfig)
	}
	o.envConfigs = envConfigs

	return nil
}

func (o *initPipelineOpts) askRepository() error {
	var err error
	if o.repoURL == "" {
		if err = o.selectURL(); err != nil {
			return err
		}
	}

	if strings.Contains(o.repoURL, githubURL) {
		o.provider = ghProviderName
		if o.githubOwner, o.repoName, err = o.parseOwnerRepoName(o.repoURL); err != nil {
			return err
		}
		if o.githubAccessToken == "" {
			if err = o.getGitHubAccessToken(); err != nil {
				return err
			}
		}
		if o.repoBranch == "" {
			o.repoBranch = defaultGHBranch
		}
	}
	if strings.Contains(o.repoURL, codeCommit) {
		o.provider = ccProviderName
		if o.repoName, err = o.parseRepoName(o.repoURL); err != nil {
			return err
		}
		if o.ccRegion, err = o.parseRegion(o.repoURL); err != nil {
			return err
		}

		// If any one of the chosen environments is in a region besides that of the CodeCommit repo, pipeline init errors out.
		for _, env := range o.envConfigs {
			if env.Region != o.ccRegion {
				return fmt.Errorf("repository %s is in %s, but environment %s is in %s; they must be in the same region", o.repoName, o.ccRegion, env.Name, env.Region)
			}
		}
		if o.repoBranch == "" {
			o.repoBranch = defaultCCBranch
		}
	}
	return nil
}

func (o *initPipelineOpts) selectURL() error {
	// Fetches and parses all remote repositories.
	err := o.runner.Run("git", []string{"remote", "-v"}, command.Stdout(&o.buffer))
	if err != nil {
		return fmt.Errorf("get remote repository info: %w; make sure you have installed Git and are in a Git repository", err)
	}
	urls, err := o.parseGitRemoteResult(strings.TrimSpace(o.buffer.String()))
	if err != nil {
		return err
	}
	o.buffer.Reset()

	// Prompts user to select a repo URL.
	url, err := o.prompt.SelectOne(
		pipelineSelectURLPrompt,
		pipelineSelectURLHelpPrompt,
		urls,
	)
	if err != nil {
		return fmt.Errorf("select URL: %w", err)
	}
	if err := o.validateURL(url); err != nil {
		return err
	}
	o.repoURL = url

	return nil

}

func (o *initPipelineOpts) parseOwnerRepoName(url string) (string, string, error) {
	regexPattern := regexp.MustCompile(`.*(github.com)(:|\/)`)
	parsedURL := strings.TrimPrefix(url, regexPattern.FindString(url))
	parsedURL = strings.TrimSuffix(parsedURL, ".git")
	ownerRepo := strings.Split(parsedURL, "/")
	if len(ownerRepo) != 2 {
		return "", "", fmt.Errorf("unable to parse the GitHub repository owner and name from %s: please pass the repository URL with the format `--url https://github.com/{owner}/{repositoryName}`", url)
	}
	return ownerRepo[0], ownerRepo[1], nil
}

func (o *initPipelineOpts) parseRepoName(url string) (string, error) {
	parsedForRepo := strings.Split(url, "/")
	if len(parsedForRepo) < 2 {
		return "", fmt.Errorf("unable to parse the CodeCommit repository name from %s", url)
	}
	return parsedForRepo[len(parsedForRepo)-1], nil
}

func (o *initPipelineOpts) parseRegion(url string) (string, error) {
	regexPattern := regexp.MustCompile(`(codecommit)(::|.)`)
	region := regexPattern.Split(url, 2)
	region = strings.Split(region[1], ".")
	region = strings.Split(region[0], ":")
	match, _ := regexp.MatchString(`(us(-gov)?|ap|ca|cn|eu|sa)-(central|(north|south)?(east|west)?)-\d`, region[0])
	if !match {
		return "", fmt.Errorf("unable to parse the AWS region from %s", url)
	}
	return region[0], nil
}

// examples:
// efekarakus	git@github.com:efekarakus/grit.git (fetch)
// efekarakus	https://github.com/karakuse/grit.git (fetch)
// origin	    https://github.com/koke/grit (fetch)
// koke         git://github.com/koke/grit.git (push)

// https	https://git-codecommit.us-west-2.amazonaws.com/v1/repos/aws-sample (fetch)
// fed		codecommit::us-west-2://aws-sample (fetch)
// ssh		ssh://git-codecommit.us-west-2.amazonaws.com/v1/repos/aws-sample (push)
func (o *initPipelineOpts) parseGitRemoteResult(s string) ([]string, error) {
	var urls []string
	urlSet := make(map[string]bool)
	items := strings.Split(s, "\n")
	for _, item := range items {
		if !strings.Contains(item, githubURL) && !strings.Contains(item, codeCommit) {
			continue
		}
		cols := strings.Split(item, "\t")
		url := strings.TrimSpace(strings.TrimSuffix(strings.Split(cols[1], " ")[0], ".git"))
		urlSet[url] = true
	}
	for url := range urlSet {
		urls = append(urls, url)
	}
	return urls, nil
}

func (o *initPipelineOpts) getGitHubAccessToken() error {
	token, err := o.prompt.GetSecret(
		fmt.Sprintf("Please enter your GitHub Personal Access Token for your repository %s:", color.HighlightUserInput(o.repoName)),
		`The personal access token for the GitHub repository linked to your workspace. 
For more information, please refer to: https://git.io/JfDFD.`,
	)

	if err != nil {
		return fmt.Errorf("get GitHub access token: %w", err)
	}
	o.githubAccessToken = token
	return nil
}

func (o *initPipelineOpts) createPipelineManifest() error {
	pipelineName, err := o.pipelineName()
	if err != nil {
		return err
	}
	provider, err := o.pipelineProvider()
	if err != nil {
		return err
	}

	var stages []manifest.PipelineStage
	for _, env := range o.envConfigs {

		stage := manifest.PipelineStage{
			Name:             env.Name,
			RequiresApproval: env.Prod,
		}
		stages = append(stages, stage)
	}

	manifest, err := manifest.NewPipelineManifest(pipelineName, provider, stages)
	if err != nil {
		return fmt.Errorf("generate a pipeline manifest: %w", err)
	}

	var manifestExists bool
	manifestPath, err := o.workspace.WritePipelineManifest(manifest)
	if err != nil {
		e, ok := err.(*workspace.ErrFileExists)
		if !ok {
			return fmt.Errorf("write pipeline manifest to workspace: %w", err)
		}
		manifestExists = true
		manifestPath = e.FileName
	}

	manifestPath, err = relPath(manifestPath)
	if err != nil {
		return err
	}

	manifestMsgFmt := "Wrote the pipeline manifest for %s at '%s'\n"
	if manifestExists {
		manifestMsgFmt = "Pipeline manifest file for %s already exists at %s, skipping writing it.\n"
	}
	log.Successf(manifestMsgFmt, color.HighlightUserInput(o.repoName), color.HighlightResource(manifestPath))
	log.Infoln("The manifest contains configurations for your CodePipeline resources, such as your pipeline stages and build steps.")
	return nil
}

func (o *initPipelineOpts) createBuildspec() error {
	artifactBuckets, err := o.artifactBuckets()
	if err != nil {
		return err
	}
	content, err := o.parser.Parse(buildspecTemplatePath, struct {
		BinaryS3BucketPath string
		Version            string
		ArtifactBuckets    []artifactBucket
	}{
		BinaryS3BucketPath: binaryS3BucketPath,
		Version:            version.Version,
		ArtifactBuckets:    artifactBuckets,
	})
	if err != nil {
		return err
	}
	buildspecPath, err := o.workspace.WritePipelineBuildspec(content)
	var buildspecExists bool
	if err != nil {
		e, ok := err.(*workspace.ErrFileExists)
		if !ok {
			return fmt.Errorf("write buildspec to workspace: %w", err)
		}
		buildspecExists = true
		buildspecPath = e.FileName
	}
	buildspecMsgFmt := "Wrote the buildspec for the pipeline's build stage at '%s'\n"
	if buildspecExists {
		buildspecMsgFmt = "Buildspec file for pipeline already exists at %s, skipping writing it.\n"
	}
	buildspecPath, err = relPath(buildspecPath)
	if err != nil {
		return err
	}
	log.Successf(buildspecMsgFmt, color.HighlightResource(buildspecPath))
	log.Infoln("The buildspec contains the commands to build and push your container images to your ECR repositories.")

	return nil
}

func (o *initPipelineOpts) secretName() string {
	return fmt.Sprintf(fmtSecretName, o.appName, o.repoName)
}

func (o *initPipelineOpts) pipelineName() (string, error) {
	if o.provider == ghProviderName {
		return fmt.Sprintf(fmtGHPipelineName, o.appName, o.githubOwner, o.repoName), nil
	}
	if o.provider == ccProviderName {
		return fmt.Sprintf(fmtCCPipelineName, o.appName, o.repoName), nil
	}
	return "", fmt.Errorf("unable to create pipeline name for repo %s from provider %s", o.repoName, o.provider)
}

func (o *initPipelineOpts) pipelineProvider() (manifest.Provider, error) {
	if o.provider == ghProviderName {
		config := &manifest.GitHubProperties{
			OwnerAndRepository:    fmt.Sprintf(fmtGHPipelineProvider, githubURL, o.githubOwner, o.repoName),
			Branch:                o.repoBranch,
			GithubSecretIdKeyName: o.secret,
		}
		return manifest.NewProvider(config)
	}
	if o.provider == ccProviderName {
		config := &manifest.CodeCommitProperties{
			Repository: fmt.Sprintf(fmtCCPipelineProvider, o.ccRegion, awsURL, o.repoName),
			Branch:     o.repoBranch,
		}
		return manifest.NewProvider(config)
	}
	return nil, fmt.Errorf("unable to create pipeline source provider for %s", o.repoName)
}

func (o *initPipelineOpts) artifactBuckets() ([]artifactBucket, error) {
	app, err := o.store.GetApplication(o.appName)
	if err != nil {
		return nil, fmt.Errorf("get application %s: %w", o.appName, err)
	}
	regionalResources, err := o.cfnClient.GetRegionalAppResources(app)
	if err != nil {
		return nil, fmt.Errorf("get regional application resources: %w", err)
	}

	var buckets []artifactBucket
	for _, resource := range regionalResources {
		var envNames []string
		for _, env := range o.envConfigs {
			if env.Region == resource.Region {
				envNames = append(envNames, env.Name)
			}
		}
		bucket := artifactBucket{
			BucketName:   resource.S3Bucket,
			Region:       resource.Region,
			Environments: envNames,
		}
		buckets = append(buckets, bucket)
	}
	return buckets, nil
}

// buildPipelineInitCmd build the command for creating a new pipeline.
func buildPipelineInitCmd() *cobra.Command {
	vars := initPipelineVars{}
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Creates a pipeline for the services in your workspace.",
		Long:  `Creates a pipeline for the services in your workspace, using the environments associated with the application.`,
		Example: `
  Create a pipeline for the services in your workspace.
  /code $ copilot pipeline init \
  /code  --url https://github.com/gitHubUserName/myFrontendApp.git \
  /code  --github-access-token file://myGitHubToken \
  /code  --environments "stage,prod"`,
		RunE: runCmdE(func(cmd *cobra.Command, args []string) error {
			opts, err := newInitPipelineOpts(vars)
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
			log.Infoln()
			log.Infoln("Recommended follow-up actions:")
			for _, followup := range opts.RecommendedActions() {
				log.Infof("- %s\n", followup)
			}
			return nil
		}),
	}
	cmd.Flags().StringVarP(&vars.appName, appFlag, appFlagShort, tryReadingAppName(), appFlagDescription)
	cmd.Flags().StringVar(&vars.repoURL, githubURLFlag, "", githubURLFlagDescription)
	_ = cmd.Flags().MarkHidden(githubURLFlag)
	cmd.Flags().StringVarP(&vars.repoURL, repoURLFlag, repoURLFlagShort, "", repoURLFlagDescription)
	cmd.Flags().StringVarP(&vars.githubAccessToken, githubAccessTokenFlag, githubAccessTokenFlagShort, "", githubAccessTokenFlagDescription)
	cmd.Flags().StringVarP(&vars.repoBranch, gitBranchFlag, gitBranchFlagShort, "", gitBranchFlagDescription)
	cmd.Flags().StringSliceVarP(&vars.environments, envsFlag, envsFlagShort, []string{}, pipelineEnvsFlagDescription)

	return cmd
}
