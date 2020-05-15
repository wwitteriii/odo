package pipelines

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/openshift/odo/pkg/odo/cli/pipelines/scm"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/openshift/odo/pkg/pipelines"
	"github.com/openshift/odo/pkg/pipelines/ioutils"
	"github.com/spf13/cobra"

	ktemplates "k8s.io/kubernetes/pkg/kubectl/util/templates"
)

const (
	// BootstrapRecommendedCommandName the recommended command name
	BootstrapRecommendedCommandName = "bootstrap"
)

var (
	bootstrapExample = ktemplates.Examples(`
    # Bootstrap OpenShift pipelines.
    %[1]s 
    `)

	bootstrapLongDesc  = ktemplates.LongDesc(`Bootstrap GitOps CI/CD Manifest`)
	bootstrapShortDesc = `Bootstrap pipelines with a starter configuration`
)

// BootstrapParameters encapsulates the parameters for the odo pipelines init command.
type BootstrapParameters struct {
	gitOpsRepoURL            string // This is where the pipelines and configuration are.
	gitOpsWebhookSecret      string // This is the secret for authenticating hooks from your GitOps repo.
	appRepoURL               string // This is the full URL to your GitHub repository for your app source.
	appWebhookSecret         string // This is the secret for authenticating hooks from your app source.
	internalRegistryHostname string // This is the internal registry hostname used for pushing images.
	imageRepo                string // This is where built images are pushed to.
	prefix                   string // Used to prefix generated environment names in a shared cluster.
	outputPath               string // Where to write the bootstrapped files to?
	dockerConfigJSONFilename string
	// generic context options common to all commands
	*genericclioptions.Context
}

// NewBootstrapParameters bootstraps a BootstrapParameters instance.
func NewBootstrapParameters() *BootstrapParameters {
	return &BootstrapParameters{}
}

// Complete completes BootstrapParameters after they've been created.
//
// If the prefix provided doesn't have a "-" then one is added, this makes the
// generated environment names nicer to read.
func (io *BootstrapParameters) Complete(name string, cmd *cobra.Command, args []string) error {
	if io.prefix != "" && !strings.HasSuffix(io.prefix, "-") {
		io.prefix = io.prefix + "-"
	}
	return nil
}

// Validate validates the parameters of the BootstrapParameters.
func (io *BootstrapParameters) Validate() error {
	gr, err := url.Parse(io.gitOpsRepoURL)
	if err != nil {
		return fmt.Errorf("failed to parse url %s: %w", io.gitOpsRepoURL, err)
	}

	// TODO: this won't work with GitLab as the repo can have more path elements.
	if len(removeEmptyStrings(strings.Split(gr.Path, "/"))) != 2 {
		return fmt.Errorf("repo must be org/repo: %s", strings.Trim(gr.Path, ".git"))
	}
	return nil
}

// Run runs the project bootstrap command.
func (io *BootstrapParameters) Run() error {
	gitOpsRep, err := scm.NewRepository(io.gitOpsRepoURL)
	if err != nil {
		return err
	}
	svcRepo, err := scm.NewRepository(io.appRepoURL)
	if err != nil {
		return err
	}
	options := pipelines.BootstrapOptions{
		GitOpsRepo:               gitOpsRep,
		AppRepo:                  svcRepo,
		GitOpsWebhookSecret:      io.gitOpsWebhookSecret,
		AppWebhookSecret:         io.appWebhookSecret,
		ImageRepo:                io.imageRepo,
		InternalRegistryHostname: io.internalRegistryHostname,
		Prefix:                   io.prefix,
		OutputPath:               io.outputPath,
		DockerConfigJSONFilename: io.dockerConfigJSONFilename,
	}
	return pipelines.Bootstrap(&options, ioutils.NewFilesystem())
}

// NewCmdBootstrap creates the project init command.
func NewCmdBootstrap(name, fullName string) *cobra.Command {
	o := NewBootstrapParameters()

	initCmd := &cobra.Command{
		Use:     name,
		Short:   initShortDesc,
		Long:    initLongDesc,
		Example: fmt.Sprintf(initExample, fullName),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}

	initCmd.Flags().StringVar(&o.gitOpsRepoURL, "gitops-repo-url", "", "GitOps repository e.g. https://github.com/organisation/repository")
	initCmd.Flags().StringVar(&o.gitOpsWebhookSecret, "gitops-webhook-secret", "", "provide the GitHub webhook secret for GitOps repository")

	initCmd.Flags().StringVar(&o.appRepoURL, "app-repo-url", "", "Application source e.g. https://github.com/organisation/application")
	initCmd.Flags().StringVar(&o.appWebhookSecret, "app-webhook-secret", "", "Provide the GitHub webhook secret for Application repository")

	initCmd.Flags().StringVar(&o.dockerConfigJSONFilename, "dockercfgjson", "", "provide the dockercfgjson path")
	initCmd.Flags().StringVar(&o.internalRegistryHostname, "internal-registry-hostname", "image-registry.openshift-image-registry.svc:5000", "internal image registry hostname")
	initCmd.Flags().StringVar(&o.outputPath, "output", ".", "folder path to add Gitops resources")
	initCmd.Flags().StringVarP(&o.prefix, "prefix", "p", "", "add a prefix to the environment names")
	initCmd.Flags().StringVarP(&o.imageRepo, "image-repo", "", "", "used to push built images")

	initCmd.MarkFlagRequired("gitops-repo-url")
	initCmd.MarkFlagRequired("gitops-webhook-secret")
	initCmd.MarkFlagRequired("app-repo-url")
	initCmd.MarkFlagRequired("app-webhook-secret")
	initCmd.MarkFlagRequired("dockercfgjson")
	initCmd.MarkFlagRequired("image-repo")

	return initCmd
}

func removeEmptyStrings(s []string) []string {
	nonempty := []string{}
	for _, v := range s {
		if v != "" {
			nonempty = append(nonempty, v)
		}
	}
	return nonempty
}
