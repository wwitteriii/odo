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
	io.GitOpsRepoURL = ui.EnterInteractiveCommandLineGitRepo()
	option := ui.SelectOptionImageRepository()
	if option == "Openshift Internal repository" {
		io.InternalRegistryHostname = ui.EnterInteractiveCommandLineInternalRegistry()
		io.ImageRepo = ui.EnterInteractiveCommandLineImageRepoInternalRegistry()

	} else {
		io.DockerConfigJSONFilename = ui.EnterInteractiveCommandLineDockercfg()
		io.ImageRepo = ui.EnterInteractiveCommandLineImageRepoExternalRepository()
	}
	io.OutputPath = ui.EnterInteractiveCommandLineOutputPath()
	exists, _ := ioutils.IsExisting(ioutils.NewFilesystem(), filepath.Join(io.OutputPath, "pipelines.yaml"))

	if !exists {
		io.Overwrite = true
	} else {
		selectOverwriteOption := ui.SelectOptionOverwrite()
		if selectOverwriteOption == "no" {
			io.Overwrite = false
			return fmt.Errorf("Cannot create Gitops configuration since file exists at %s", io.OutputPath)
		}

		io.Overwrite = true
	}
	io.GitOpsWebhookSecret = ui.EnterInteractiveCommandLineGitWebhookSecret()
	io.SealedSecretsService.Name = ui.EnterInteractiveCommandLineSealedSecrets()
	io.SealedSecretsService.Namespace = ui.EnterInteractiveCommandLineSealedSecretNamespace()
	commitStatusTrackerCheck := ui.SelectOptionCommitStatusTracker()
	if commitStatusTrackerCheck == "yes" {
		io.StatusTrackerAccessToken = ui.EnterInteractiveCommandLineStatusTrackerAccessToken()
	}
	io.Prefix = ui.EnterInteractiveCommandLinePrefix()
	io.Prefix = utility.MaybeCompletePrefix(io.Prefix)
	InitOption := ui.SelectOptionBootstrap()
	if InitOption == "Bootstrap" {
		io.ServiceRepoURL = ui.EnterInteractiveCommandLineServiceRepoURL()
		io.ServiceWebhookSecret = ui.EnterInteractiveCommandLineServiceWebhookSecret()
		io.ServiceRepoURL = utility.AddGitSuffixIfNecessary(io.ServiceRepoURL)
	}

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
