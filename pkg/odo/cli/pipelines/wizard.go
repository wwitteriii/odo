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
//
// If the prefix provided doesn't have a "-" then one is added, this makes the
// generated environment names nicer to read.
func (io *WizardParameters) Complete(name string, cmd *cobra.Command, args []string) error {
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
	io.Prefix = utility.MaybeCompletePrefix(io.Prefix)
	InitOption := ui.SelectOptionBootstrap()
	if InitOption == "Bootstrap" {
		io.ServiceRepoURL = ui.EnterServiceRepoURL()
		io.ServiceWebhookSecret = ui.EnterServiceWebhookSecret()
		io.ServiceRepoURL = utility.AddGitSuffixIfNecessary(io.ServiceRepoURL)
	}

	io.OutputPath = ui.EnterOutputPath()
	exists, _ := ioutils.IsExisting(ioutils.NewFilesystem(), filepath.Join(io.OutputPath, "pipelines.yaml"))
	if exists {
		selectOverwriteOption := ui.SelectOptionOverwrite()
		if selectOverwriteOption == "no" {
			io.Overwrite = false
			return fmt.Errorf("Cannot create GitOps configuration since file exists at %s", io.OutputPath)
		}
	}
	io.Overwrite = true
	io.GitOpsRepoURL = utility.AddGitSuffixIfNecessary(io.GitOpsRepoURL)
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
	return wizardCmd
}
