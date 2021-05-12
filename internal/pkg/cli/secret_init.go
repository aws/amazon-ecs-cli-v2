// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"errors"
	"fmt"

	termprogress "github.com/aws/copilot-cli/internal/pkg/term/progress"

	"gopkg.in/yaml.v3"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/aws/copilot-cli/internal/pkg/aws/sessions"
	"github.com/aws/copilot-cli/internal/pkg/aws/ssm"
	"github.com/aws/copilot-cli/internal/pkg/config"
	"github.com/aws/copilot-cli/internal/pkg/deploy"
	"github.com/aws/copilot-cli/internal/pkg/term/color"
	"github.com/aws/copilot-cli/internal/pkg/term/log"
	"github.com/aws/copilot-cli/internal/pkg/term/prompt"
	"github.com/aws/copilot-cli/internal/pkg/term/selector"
)

const (
	fmtSecretParameterName = "/copilot/%s/%s/secrets/%s"
)

const (
	secretInitAppPrompt     = "Which application do you want to add the secret to?"
	secretInitAppPromptHelp = "The secret can then be versioned by your existing environments inside the application."

	secretInitSecretNamePrompt     = "What would you like to name this secret?"
	secretInitSecretNamePromptHelp = "The name of the secret, such as 'db_password'."

	fmtSecretInitSecretValuePrompt     = "What is the value of secret %s in environment %s?"
	fmtSecretInitSecretValuePromptHelp = "If you do not wish to add the secret %s to environment %s, you can leave this blank by pressing 'Enter' without entering any value."
)

type secretInitVars struct {
	appName string

	name          string
	values        map[string]string
	inputFilePath string
	overwrite     bool

	resourceTags map[string]string
}

type secretInitOpts struct {
	secretInitVars

	store store
	fs    afero.Fs

	prompter prompter
	selector appSelector

	configureSecretPutter func(envName string) (secretPutter, error)
	readFile              func() ([]byte, error)
}

func newSecretInitOpts(vars secretInitVars) (*secretInitOpts, error) {
	store, err := config.NewStore()
	if err != nil {
		return nil, fmt.Errorf("new config store: %w", err)
	}

	prompter := prompt.New()
	opts := secretInitOpts{
		secretInitVars: vars,
		store:          store,
		fs:             &afero.Afero{Fs: afero.NewOsFs()},

		prompter: prompter,
		selector: selector.NewSelect(prompter, store),
	}

	opts.configureSecretPutter = func(envName string) (secretPutter, error) {
		env, err := opts.targetEnv(envName)
		if err != nil {
			return nil, err
		}
		sess, err := sessions.NewProvider().FromRole(env.ManagerRoleARN, env.Region)
		if err != nil {
			return nil, fmt.Errorf("create session from environment manager role %s in region %s: %w", env.ManagerRoleARN, env.Region, err)
		}
		return ssm.New(sess), nil
	}

	opts.readFile = func() ([]byte, error) {
		file, err := opts.fs.Open(opts.inputFilePath)
		if err != nil {
			return nil, fmt.Errorf("open input file %s: %w", opts.inputFilePath, err)
		}
		defer file.Close()
		f, err := afero.ReadFile(opts.fs, file.Name())
		if err != nil {
			return nil, fmt.Errorf("read input file %s: %w", opts.inputFilePath, err)
		}

		return f, nil
	}
	return &opts, nil
}

// Validate returns an error if the flag values passed by the user are invalid.
func (o *secretInitOpts) Validate() error {
	if o.inputFilePath != "" && o.name != "" {
		return errors.New("cannot specify `--cli-input-yaml` with `--name`")
	}

	if o.inputFilePath != "" && o.values != nil {
		return errors.New("cannot specify `--cli-input-yaml` with `--values`")
	}

	if o.appName != "" {
		_, err := o.store.GetApplication(o.appName)
		if err != nil {
			return fmt.Errorf("get application %s: %w", o.appName, err)
		}
	}

	if o.name != "" {
		if err := validateSecretName(o.name); err != nil {
			return err
		}
	}

	if o.values != nil {
		for env := range o.values {
			if _, err := o.targetEnv(env); err != nil {
				return err
			}
		}
	}

	if o.inputFilePath != "" {
		if _, err := o.fs.Stat(o.inputFilePath); err != nil {
			return err
		}
	}
	return nil
}

// Ask prompts the user for any required or important fields that are not provided.
func (o *secretInitOpts) Ask() error {
	if o.overwrite {
		log.Warningf("You have specified %s flag. Please note that overwriting an existing secret may break your deployed service.\n", color.HighlightCode("--overwrite"))
	}

	if o.inputFilePath != "" {
		return nil
	}

	if err := o.askForAppName(); err != nil {
		return err
	}
	if err := o.askForSecretName(); err != nil {
		return err
	}
	if err := o.askForSecretValues(); err != nil {
		return err
	}
	return nil
}

// Execute creates or updates the secrets.
func (o *secretInitOpts) Execute() error {
	if o.inputFilePath != "" {
		secrets, err := o.parseSecretsInputFile()
		if err != nil {
			return err
		}
		for secretName, secretValues := range secrets {
			o.putSecret(secretName, secretValues)
		}
		return nil
	}

	o.putSecret(o.name, o.values)
	return nil
}

func (o *secretInitOpts) putSecret(secretName string, values map[string]string) {
	spinner := termprogress.NewSpinner(log.DiagnosticWriter)

	spinner.Start("start creating secret")
	errLogs := make([]string, 0)
	for envName, value := range values {
		err := o.putSecretInEnv(secretName, envName, value)
		if err != nil {
			errLogs = append(errLogs, fmt.Sprintf("Failed to put secret %s in environment %s: %s", color.HighlightUserInput(secretName), color.HighlightUserInput(envName), err.Error()))
			continue
		}
	}

	for _, errLog := range errLogs {
		log.Errorln(errLog)
	}
	spinner.Stop("finish creating secret")
	log.Infoln("")
}

func (o *secretInitOpts) putSecretInEnv(secretName, envName, value string) error {
	client, err := o.configureSecretPutter(envName)
	if err != nil {
		return err
	}

	tags := make(map[string]string)
	for k, v := range o.resourceTags {
		tags[k] = v
	}
	tags[deploy.AppTagKey] = o.appName
	tags[deploy.EnvTagKey] = envName

	name := fmt.Sprintf(fmtSecretParameterName, o.appName, envName, secretName)
	in := ssm.PutSecretInput{
		Name:      name,
		Value:     value,
		Overwrite: o.overwrite,
		Tags:      tags,
	}

	out, err := client.PutSecret(in)
	if err != nil {
		var targetErr *ssm.ErrParameterAlreadyExists
		if errors.As(err, &targetErr) {
			log.Warningf("Secret %s already exists. If you want to overwrite an existing secret, use the %s flag.\n", name, color.HighlightCode("--overwrite"))
			return nil
		}
		return err
	}

	version := aws.Int64Value(out.Version)
	if version != 1 {
		log.Successln(fmt.Sprintf("Secret %s already exists in environment %s. Overwritten.", name, color.HighlightUserInput(envName)))
		return nil
	}

	log.Successln(fmt.Sprintf("Successfully put secret %s in environment %s as %s.", color.HighlightUserInput(secretName), color.HighlightUserInput(envName), color.HighlightResource(name)))
	return nil
}

func (o *secretInitOpts) parseSecretsInputFile() (map[string]map[string]string, error) {
	raw, err := o.readFile()
	if err != nil {
		return nil, err
	}

	type inputFile struct {
		Secrets map[string]map[string]string `yaml:",inline"`
	}
	var f inputFile
	if err := yaml.Unmarshal(raw, &f); err != nil {
		return nil, fmt.Errorf("unmarshal input file: %w", err)
	}
	return f.Secrets, nil
}

func (o *secretInitOpts) askForAppName() error {
	if o.appName != "" {
		return nil
	}

	app, err := o.selector.Application(secretInitAppPrompt, secretInitAppPromptHelp)
	if err != nil {
		return fmt.Errorf("ask for an application to add the secret to: %w", err)
	}
	o.appName = app
	return nil
}

func (o *secretInitOpts) askForSecretName() error {
	if o.name != "" {
		return nil
	}

	name, err := o.prompter.Get(secretInitSecretNamePrompt,
		secretInitSecretNamePromptHelp,
		validateSecretName,
		prompt.WithFinalMessage("secret name: "))
	if err != nil {
		return fmt.Errorf("ask for the secret name: %w", err)
	}

	o.name = name
	return nil
}

func (o *secretInitOpts) askForSecretValues() error {
	if o.values != nil {
		return nil
	}

	envs, err := o.store.ListEnvironments(o.appName)
	if err != nil {
		return fmt.Errorf("list environments in app %s: %w", o.appName, err)
	}

	if len(envs) == 0 {
		log.Errorf("Secrets environment-level resources. Please run %s before running %s.\n",
			color.HighlightCode("copilot env init"),
			color.HighlightCode("copilot secret init"))
		return fmt.Errorf("no environment is found in app %s", o.appName)
	}

	values := make(map[string]string)
	for _, env := range envs {
		value, err := o.prompter.GetSecret(
			fmt.Sprintf(fmtSecretInitSecretValuePrompt, color.HighlightUserInput(o.name), env.Name),
			fmt.Sprintf(fmtSecretInitSecretValuePromptHelp, color.HighlightUserInput(o.name), env.Name))
		if err != nil {
			return fmt.Errorf("get secret value for %s in environment %s: %w", color.HighlightUserInput(o.name), env.Name, err)
		}

		values[env.Name] = value
	}
	o.values = values
	return nil
}

func (o *secretInitOpts) targetEnv(envName string) (*config.Environment, error) {
	env, err := o.store.GetEnvironment(o.appName, envName)
	if err != nil {
		return nil, fmt.Errorf("get environment %s in application %s: %w", envName, o.appName, err)
	}
	return env, nil
}

// BuildSecretInitCmd build the command for creating a new secret or updating an existing one.
func BuildSecretInitCmd() *cobra.Command {
	vars := secretInitVars{}
	cmd := &cobra.Command{
		Use:     "secret init",
		Short:   "Create or update an SSM SecureString parameter.",
		Example: ``, // TODO
		RunE: runCmdE(func(cmd *cobra.Command, args []string) error {
			opts, err := newSecretInitOpts(vars)
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
			return nil
		}),
	}

	cmd.Flags().StringVar(&vars.appName, appFlag, tryReadingAppName(), appFlagDescription)
	cmd.Flags().StringVar(&vars.name, nameFlag, "", secretNameFlagDescription)
	cmd.Flags().StringToStringVar(&vars.values, valuesFlag, nil, secretValuesFlagDescription)
	cmd.Flags().BoolVar(&vars.overwrite, overwriteFlag, false, secretOverwriteFlagDescription)
	cmd.Flags().StringVar(&vars.inputFilePath, inputFilePathFlag, "", secretInputFilePathFlagDescription)
	cmd.Flags().StringToStringVar(&vars.resourceTags, resourceTagsFlag, nil, resourceTagsFlagDescription)
	return cmd
}
