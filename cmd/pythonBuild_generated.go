// Code generated by piper's step-generator. DO NOT EDIT.

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/gcp"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperenv"
	"github.com/SAP/jenkins-library/pkg/splunk"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/validation"
	"github.com/spf13/cobra"
)

type pythonBuildOptions struct {
	BuildFlags               []string `json:"buildFlags,omitempty"`
	SetupFlags               []string `json:"setupFlags,omitempty"`
	CreateBOM                bool     `json:"createBOM,omitempty"`
	Publish                  bool     `json:"publish,omitempty"`
	TargetRepositoryPassword string   `json:"targetRepositoryPassword,omitempty"`
	TargetRepositoryUser     string   `json:"targetRepositoryUser,omitempty"`
	TargetRepositoryURL      string   `json:"targetRepositoryURL,omitempty"`
	BuildSettingsInfo        string   `json:"buildSettingsInfo,omitempty"`
	VirutalEnvironmentName   string   `json:"virutalEnvironmentName,omitempty"`
	RequirementsFilePath     string   `json:"requirementsFilePath,omitempty"`
}

type pythonBuildCommonPipelineEnvironment struct {
	custom struct {
		buildSettingsInfo string
	}
}

func (p *pythonBuildCommonPipelineEnvironment) persist(path, resourceName string) {
	content := []struct {
		category string
		name     string
		value    interface{}
	}{
		{category: "custom", name: "buildSettingsInfo", value: p.custom.buildSettingsInfo},
	}

	errCount := 0
	for _, param := range content {
		err := piperenv.SetResourceParameter(path, resourceName, filepath.Join(param.category, param.name), param.value)
		if err != nil {
			log.Entry().WithError(err).Error("Error persisting piper environment.")
			errCount++
		}
	}
	if errCount > 0 {
		log.Entry().Error("failed to persist Piper environment")
	}
}

// PythonBuildCommand Step builds a python project
func PythonBuildCommand() *cobra.Command {
	const STEP_NAME = "pythonBuild"

	metadata := pythonBuildMetadata()
	var stepConfig pythonBuildOptions
	var startTime time.Time
	var commonPipelineEnvironment pythonBuildCommonPipelineEnvironment
	var logCollector *log.CollectorHook
	var splunkClient *splunk.Splunk
	telemetryClient := &telemetry.Telemetry{}

	var createPythonBuildCmd = &cobra.Command{
		Use:   STEP_NAME,
		Short: "Step builds a python project",
		Long: `Step build python project using the setup.py manifest and builds a wheel and tarball artifact . please note that currently python build only supports setup.py

### build with depedencies from a private repository
if your build has dependencies from a private repository you can include the standard requirements.txt into the source code with ` + "`" + `--extra-index-url` + "`" + ` as the first line

` + "`" + `` + "`" + `` + "`" + `
--extra-index-url https://${PIPER_VAULTCREDENTIAL_USERNAME}:${PIPER_VAULTCREDENTIAL_PASSWORD}@<privateRepoUrl>/simple
` + "`" + `` + "`" + `` + "`" + `
` + "`" + `PIPER_VAULTCREDENTIAL_USERNAME` + "`" + ` and ` + "`" + `PIPER_VAULTCREDENTIAL_PASSWORD` + "`" + ` are the username and password for the private repository
and are exposed are environment variables that must be present in the environment where the Piper step runs or alternatively can be created using :
[vault general purpose credentials](../infrastructure/vault.md#using-vault-for-general-purpose-and-test-credentials)`,
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			startTime = time.Now()
			log.SetStepName(STEP_NAME)
			log.SetVerbose(GeneralConfig.Verbose)

			GeneralConfig.GitHubAccessTokens = ResolveAccessTokens(GeneralConfig.GitHubTokens)

			path, err := os.Getwd()
			if err != nil {
				return err
			}
			fatalHook := &log.FatalHook{CorrelationID: GeneralConfig.CorrelationID, Path: path}
			log.RegisterHook(fatalHook)

			err = PrepareConfig(cmd, &metadata, STEP_NAME, &stepConfig, config.OpenPiperFile)
			if err != nil {
				log.SetErrorCategory(log.ErrorConfiguration)
				return err
			}
			log.RegisterSecret(stepConfig.TargetRepositoryPassword)
			log.RegisterSecret(stepConfig.TargetRepositoryUser)

			if len(GeneralConfig.HookConfig.SentryConfig.Dsn) > 0 {
				sentryHook := log.NewSentryHook(GeneralConfig.HookConfig.SentryConfig.Dsn, GeneralConfig.CorrelationID)
				log.RegisterHook(&sentryHook)
			}

			if len(GeneralConfig.HookConfig.SplunkConfig.Dsn) > 0 || len(GeneralConfig.HookConfig.SplunkConfig.ProdCriblEndpoint) > 0 {
				splunkClient = &splunk.Splunk{}
				logCollector = &log.CollectorHook{CorrelationID: GeneralConfig.CorrelationID}
				log.RegisterHook(logCollector)
			}

			if err = log.RegisterANSHookIfConfigured(GeneralConfig.CorrelationID); err != nil {
				log.Entry().WithError(err).Warn("failed to set up SAP Alert Notification Service log hook")
			}

			validation, err := validation.New(validation.WithJSONNamesForStructFields(), validation.WithPredefinedErrorMessages())
			if err != nil {
				return err
			}
			if err = validation.ValidateStruct(stepConfig); err != nil {
				log.SetErrorCategory(log.ErrorConfiguration)
				return err
			}

			return nil
		},
		Run: func(_ *cobra.Command, _ []string) {
			vaultClient := config.GlobalVaultClient()
			if vaultClient != nil {
				defer vaultClient.MustRevokeToken()
			}

			stepTelemetryData := telemetry.CustomData{}
			stepTelemetryData.ErrorCode = "1"
			handler := func() {
				commonPipelineEnvironment.persist(GeneralConfig.EnvRootPath, "commonPipelineEnvironment")
				config.RemoveVaultSecretFiles()
				stepTelemetryData.Duration = fmt.Sprintf("%v", time.Since(startTime).Milliseconds())
				stepTelemetryData.ErrorCategory = log.GetErrorCategory().String()
				stepTelemetryData.PiperCommitHash = GitCommit
				telemetryClient.SetData(&stepTelemetryData)
				telemetryClient.LogStepTelemetryData()
				if len(GeneralConfig.HookConfig.SplunkConfig.Dsn) > 0 {
					splunkClient.Initialize(GeneralConfig.CorrelationID,
						GeneralConfig.HookConfig.SplunkConfig.Dsn,
						GeneralConfig.HookConfig.SplunkConfig.Token,
						GeneralConfig.HookConfig.SplunkConfig.Index,
						GeneralConfig.HookConfig.SplunkConfig.SendLogs)
					splunkClient.Send(telemetryClient.GetData(), logCollector)
				}
				if len(GeneralConfig.HookConfig.SplunkConfig.ProdCriblEndpoint) > 0 {
					splunkClient.Initialize(GeneralConfig.CorrelationID,
						GeneralConfig.HookConfig.SplunkConfig.ProdCriblEndpoint,
						GeneralConfig.HookConfig.SplunkConfig.ProdCriblToken,
						GeneralConfig.HookConfig.SplunkConfig.ProdCriblIndex,
						GeneralConfig.HookConfig.SplunkConfig.SendLogs)
					splunkClient.Send(telemetryClient.GetData(), logCollector)
				}
				if GeneralConfig.HookConfig.GCPPubSubConfig.Enabled {
					err := gcp.NewGcpPubsubClient(
						vaultClient,
						GeneralConfig.HookConfig.GCPPubSubConfig.ProjectNumber,
						GeneralConfig.HookConfig.GCPPubSubConfig.IdentityPool,
						GeneralConfig.HookConfig.GCPPubSubConfig.IdentityProvider,
						GeneralConfig.CorrelationID,
						GeneralConfig.HookConfig.OIDCConfig.RoleID,
					).Publish(GeneralConfig.HookConfig.GCPPubSubConfig.Topic, telemetryClient.GetDataBytes())
					if err != nil {
						log.Entry().WithError(err).Warn("event publish failed")
					}
				}
			}
			log.DeferExitHandler(handler)
			defer handler()
			telemetryClient.Initialize(STEP_NAME)
			pythonBuild(stepConfig, &stepTelemetryData, &commonPipelineEnvironment)
			stepTelemetryData.ErrorCode = "0"
			log.Entry().Info("SUCCESS")
		},
	}

	addPythonBuildFlags(createPythonBuildCmd, &stepConfig)
	return createPythonBuildCmd
}

func addPythonBuildFlags(cmd *cobra.Command, stepConfig *pythonBuildOptions) {
	cmd.Flags().StringSliceVar(&stepConfig.BuildFlags, "buildFlags", []string{}, "Defines list of build flags passed to python binary.")
	cmd.Flags().StringSliceVar(&stepConfig.SetupFlags, "setupFlags", []string{}, "Defines list of flags passed to setup.py.")
	cmd.Flags().BoolVar(&stepConfig.CreateBOM, "createBOM", false, "Creates the bill of materials (BOM) using CycloneDX plugin.")
	cmd.Flags().BoolVar(&stepConfig.Publish, "publish", false, "Configures the build to publish artifacts to a repository.")
	cmd.Flags().StringVar(&stepConfig.TargetRepositoryPassword, "targetRepositoryPassword", os.Getenv("PIPER_targetRepositoryPassword"), "Password for the target repository where the compiled binaries shall be uploaded - typically provided by the CI/CD environment.")
	cmd.Flags().StringVar(&stepConfig.TargetRepositoryUser, "targetRepositoryUser", os.Getenv("PIPER_targetRepositoryUser"), "Username for the target repository where the compiled binaries shall be uploaded - typically provided by the CI/CD environment.")
	cmd.Flags().StringVar(&stepConfig.TargetRepositoryURL, "targetRepositoryURL", os.Getenv("PIPER_targetRepositoryURL"), "URL of the target repository where the compiled binaries shall be uploaded - typically provided by the CI/CD environment.")
	cmd.Flags().StringVar(&stepConfig.BuildSettingsInfo, "buildSettingsInfo", os.Getenv("PIPER_buildSettingsInfo"), "build settings info is typically filled by the step automatically to create information about the build settings that were used during the maven build . This information is typically used for compliance related processes.")
	cmd.Flags().StringVar(&stepConfig.VirutalEnvironmentName, "virutalEnvironmentName", `piperBuild-env`, "name of the virtual environment that will be used for the build")
	cmd.Flags().StringVar(&stepConfig.RequirementsFilePath, "requirementsFilePath", `requirements.txt`, "file path to the requirements.txt file needed for the sbom cycloneDx file creation.")

}

// retrieve step metadata
func pythonBuildMetadata() config.StepData {
	var theMetaData = config.StepData{
		Metadata: config.StepMetadata{
			Name:        "pythonBuild",
			Aliases:     []config.Alias{},
			Description: "Step builds a python project",
		},
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Parameters: []config.StepParameters{
					{
						Name:        "buildFlags",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "[]string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     []string{},
					},
					{
						Name:        "setupFlags",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "[]string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     []string{},
					},
					{
						Name:        "createBOM",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "STEPS", "STAGES", "PARAMETERS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     false,
					},
					{
						Name:        "publish",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"STEPS", "STAGES", "PARAMETERS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     false,
					},
					{
						Name: "targetRepositoryPassword",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "custom/repositoryPassword",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_targetRepositoryPassword"),
					},
					{
						Name: "targetRepositoryUser",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "custom/repositoryUsername",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_targetRepositoryUser"),
					},
					{
						Name: "targetRepositoryURL",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "custom/repositoryUrl",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_targetRepositoryURL"),
					},
					{
						Name: "buildSettingsInfo",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "custom/buildSettingsInfo",
							},
						},
						Scope:     []string{"STEPS", "STAGES", "PARAMETERS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_buildSettingsInfo"),
					},
					{
						Name:        "virutalEnvironmentName",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"STEPS", "STAGES", "PARAMETERS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     `piperBuild-env`,
					},
					{
						Name:        "requirementsFilePath",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"STEPS", "STAGES", "PARAMETERS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     `requirements.txt`,
					},
				},
			},
			Containers: []config.Container{
				{Name: "python", Image: "python:3.9"},
			},
			Outputs: config.StepOutputs{
				Resources: []config.StepResources{
					{
						Name: "commonPipelineEnvironment",
						Type: "piperEnvironment",
						Parameters: []map[string]interface{}{
							{"name": "custom/buildSettingsInfo"},
						},
					},
				},
			},
		},
	}
	return theMetaData
}
