package pipelines

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/odo/cli/pipelines/ui"
	"github.com/openshift/odo/pkg/odo/cli/pipelines/utility"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/openshift/odo/pkg/pipelines"
	"github.com/openshift/odo/pkg/pipelines/ioutils"
	"github.com/spf13/cobra"

	ktemplates "k8s.io/kubectl/pkg/util/templates"
)

const (
	// WizardRecommendedCommandName the recommended command name
	WizardRecommendedCommandName = "wizard"
)

var (
	WizardExample = ktemplates.Examples(`
    # Wizard OpenShift pipelines.
    %[1]s 
    `)

	WizardLongDesc  = ktemplates.LongDesc(`Wizard GitOps CI/CD Manifest`)
	WizardShortDesc = `Wizard pipelines with a starter configuration`
)

// WizardParameters encapsulates the parameters for the odo pipelines init command.
type WizardParameters struct {
	*pipelines.BootstrapOptions
	// generic context options common to all commands
	*genericclioptions.Context
}

// NewWizardParameters Wizards a WizardParameters instance.
func NewWizardParameters() *WizardParameters {
	return &WizardParameters{
		BootstrapOptions: &pipelines.BootstrapOptions{
			InitOptions: &pipelines.InitOptions{},
		},
	}
}

// Complete completes WizardParameters after they've been created.
// If the prefix provided doesn't have a "-" then one is added, this makes the
// generated environment names nicer to read.
func (io *WizardParameters) Complete(name string, cmd *cobra.Command, args []string) error {
	flagset := cmd.Flags()
	fmt.Print(flagset.NFlag(), flagset.NArg())
	if flagset.NFlag() == 0 {
		io.GitOpsRepoURL = ui.EnterGitRepo()
		option := ui.SelectOptionImageRepository()
		if option == "Openshift Internal repository" {
			io.InternalRegistryHostname = ui.EnterInternalRegistry()
			io.ImageRepo = ui.EnterImageRepoInternalRegistry()

		} else {
			io.DockerConfigJSONFilename = ui.EnterDockercfg()
			io.ImageRepo = ui.EnterImageRepoExternalRepository()
		}
		io.GitOpsWebhookSecret = ui.EnterGitWebhookSecret()
		io.SealedSecretsService.Name = ui.EnterSealedSecretService()
		io.SealedSecretsService.Namespace = ui.EnterSealedSecretNamespace()
		commitStatusTrackerCheck := ui.SelectOptionCommitStatusTracker()
		if commitStatusTrackerCheck == "yes" {
			io.StatusTrackerAccessToken = ui.EnterStatusTrackerAccessToken()
		}
		io.Prefix = ui.EnterPrefix()
		io.ServiceRepoURL = ui.EnterServiceRepoURL()
		if io.ServiceRepoURL != "" {
			io.ServiceWebhookSecret = ui.EnterServiceWebhookSecret()
		}

		io.OutputPath = ui.EnterOutputPath(io.GitOpsRepoURL)
		exists, _ := ioutils.IsExisting(ioutils.NewFilesystem(), filepath.Join(io.OutputPath, "pipelines.yaml"))
		if exists {
			selectOverwriteOption := ui.SelectOptionOverwrite()
			if selectOverwriteOption == "no" {
				io.Overwrite = false
				return fmt.Errorf("Cannot create GitOps configuration since file exists at %s", io.OutputPath)
			}
		}
	}
	io.Overwrite = true
	io.Prefix = utility.MaybeCompletePrefix(io.Prefix)
	io.GitOpsRepoURL = utility.AddGitSuffixIfNecessary(io.GitOpsRepoURL)
	io.ServiceRepoURL = utility.AddGitSuffixIfNecessary(io.ServiceRepoURL)
	return nil
}

// Validate validates the parameters of the WizardParameters.
func (io *WizardParameters) Validate() error {
	gr, err := url.Parse(io.GitOpsRepoURL)
	if err != nil {
		return fmt.Errorf("failed to parse url %s: %w", io.GitOpsRepoURL, err)
	}
	// TODO: this won't work with GitLab as the repo can have more path elements.
	if len(utility.RemoveEmptyStrings(strings.Split(gr.Path, "/"))) != 2 {
		return fmt.Errorf("repo must be org/repo: %s", strings.Trim(gr.Path, ".git"))
	}
	return nil
}

// Run runs the project Wizard command.
func (io *WizardParameters) Run() error {
	if io.ServiceRepoURL != "" {
		err := pipelines.Bootstrap(io.BootstrapOptions, ioutils.NewFilesystem())
		if err != nil {
			return err
		}
		log.Success("Bootstrapped GitOps sucessfully.")
	} else {
		err := pipelines.Init(io.InitOptions, ioutils.NewFilesystem())
		if err != nil {
			return err
		}
		log.Success("Initialized GitOps sucessfully.")
	}
	return nil
}

// NewCmdwizard creates the project init command.
func NewCmdWizard(name, fullName string) *cobra.Command {
	o := NewWizardParameters()

	wizardCmd := &cobra.Command{
		Use:     name,
		Short:   WizardShortDesc,
		Long:    WizardLongDesc,
		Example: fmt.Sprintf(WizardExample, fullName),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}
	addInitCommands(wizardCmd, o.BootstrapOptions.InitOptions)
	wizardCmd.Flags().StringVar(&o.ServiceRepoURL, "service-repo-url", "", "Provide the URL for your Service repository e.g. https://github.com/organisation/service.git")
	wizardCmd.Flags().StringVar(&o.ServiceWebhookSecret, "service-webhook-secret", "", "Provide a secret that we can use to authenticate incoming hooks from your Git hosting service for the Service repository. (if not provided, it will be auto-generated)")
	// wizardCmd.MarkFlagRequired("gitops-repo-url")

	return wizardCmd
}
