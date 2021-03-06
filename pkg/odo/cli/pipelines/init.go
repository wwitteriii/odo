package pipelines

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/odo/cli/pipelines/utility"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/openshift/odo/pkg/pipelines"
	"github.com/openshift/odo/pkg/pipelines/ioutils"

	"github.com/spf13/cobra"

	ktemplates "k8s.io/kubectl/pkg/util/templates"
)

const (
	// InitRecommendedCommandName the recommended command name
	InitRecommendedCommandName = "init"
)

var (
	initExample = ktemplates.Examples(`
	# Initialize OpenShift GitOps pipelines
	%[1]s 
	`)

	initLongDesc  = ktemplates.LongDesc(`Initialize GitOps pipelines`)
	initShortDesc = `Initialize pipelines`
)

// InitParameters encapsulates the parameters for the odo pipelines init command.
type InitParameters struct {
	*pipelines.InitOptions
	// generic context options common to all commands
	*genericclioptions.Context
}

// NewInitParameters bootstraps a InitParameters instance.
func NewInitParameters() *InitParameters {
	return &InitParameters{InitOptions: &pipelines.InitOptions{}}
}

// Complete completes InitParameters after they've been created.
//
// If the prefix provided doesn't have a "-" then one is added, this makes the
// generated environment names nicer to read.
func (io *InitParameters) Complete(name string, cmd *cobra.Command, args []string) error {
	io.Prefix = utility.MaybeCompletePrefix(io.Prefix)
	io.GitOpsRepoURL = utility.AddGitSuffixIfNecessary(io.GitOpsRepoURL)
	return nil
}

// Validate validates the parameters of the InitParameters.
func (io *InitParameters) Validate() error {
	gr, err := url.Parse(io.GitOpsRepoURL)
	if err != nil {
		return fmt.Errorf("failed to parse url %s: %v", io.GitOpsRepoURL, err)
	}

	// TODO: this won't work with GitLab as the repo can have more path elements.
	if len(utility.RemoveEmptyStrings(strings.Split(gr.Path, "/"))) != 2 {
		return fmt.Errorf("repo must be org/repo: %s", strings.Trim(gr.Path, ".git"))
	}

	return nil
}

// Run runs the project bootstrap command.
func (io *InitParameters) Run() error {
	err := pipelines.Init(io.InitOptions, ioutils.NewFilesystem())
	if err != nil {
		return err
	}
	log.Successf("Intialized GitOps sucessfully.")
	return nil
}

// NewCmdInit creates the project init command.
func NewCmdInit(name, fullName string) *cobra.Command {
	o := NewInitParameters()

	initCmd := &cobra.Command{
		Use:     name,
		Short:   initShortDesc,
		Long:    initLongDesc,
		Example: fmt.Sprintf(initExample, fullName),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}

	addInitCommands(initCmd, o.InitOptions)

	initCmd.MarkFlagRequired("gitops-repo-url")
	return initCmd
}

func addInitCommands(cmd *cobra.Command, o *pipelines.InitOptions) {
	cmd.Flags().StringVar(&o.GitOpsRepoURL, "gitops-repo-url", "", "Provide the URL for your GitOps repository e.g. https://github.com/organisation/repository.git")
	cmd.Flags().StringVar(&o.GitOpsWebhookSecret, "gitops-webhook-secret", "", "Provide a secret that we can use to authenticate incoming hooks from your Git hosting service for the GitOps repository. (if not provided, it will be auto-generated)")
	cmd.Flags().StringVar(&o.OutputPath, "output", ".", "Path to write GitOps resources")
	cmd.Flags().StringVarP(&o.Prefix, "prefix", "p", "", "Add a prefix to the environment names(Dev, stage,prod,cicd etc.) to distinguish and identify individual environments")
	cmd.Flags().StringVar(&o.DockerConfigJSONFilename, "dockercfgjson", "~/.docker/config.json", "Filepath to config.json which authenticates the image push to the desired image registry ")
	cmd.Flags().StringVar(&o.InternalRegistryHostname, "image-repo-internal-registry-hostname", "image-registry.openshift-image-registry.svc:5000", "Host-name for internal image registry e.g. docker-registry.default.svc.cluster.local:5000, used if you are pushing your images to the internal image registry")
	cmd.Flags().StringVar(&o.ImageRepo, "image-repo", "", "Image repository of the form <registry>/<username>/<repository> or <project>/<app> which is used to push newly built images")
	cmd.Flags().StringVar(&o.SealedSecretsService.Namespace, "sealed-secrets-ns", "kube-system", "Namespace in which the Sealed Secrets operator is installed, automatically generated secrets are encrypted with this operator")
	cmd.Flags().StringVar(&o.SealedSecretsService.Name, "sealed-secrets-svc", "sealed-secrets-controller", "Name of the Sealed Secrets Services that encrypts secrets")
	cmd.Flags().StringVar(&o.StatusTrackerAccessToken, "status-tracker-access-token", "", "Used to authenticate requests to push commit-statuses to your Git hosting service")
	cmd.Flags().BoolVar(&o.Overwrite, "overwrite", false, "Overwrites previously existing GitOps configuration (if any)")

	cmd.MarkFlagRequired("gitops-repo-url")
}
