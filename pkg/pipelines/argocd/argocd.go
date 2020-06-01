package argocd

import (
	"path/filepath"
	"sort"

	// This is a hack because ArgoCD doesn't support a compatible (code-wise)
	// version of k8s in common with odo.
	argoappv1 "github.com/openshift/odo/pkg/pipelines/argocd/v1alpha1"

	"github.com/openshift/odo/pkg/pipelines/config"
	"github.com/openshift/odo/pkg/pipelines/meta"
	res "github.com/openshift/odo/pkg/pipelines/resources"
)

var (
	applicationTypeMeta = meta.TypeMeta(
		"Application",
		"argoproj.io/v1alpha1",
	)

	syncPolicy = &argoappv1.SyncPolicy{
		Automated: &argoappv1.SyncPolicyAutomated{
			Prune:    true,
			SelfHeal: true,
		},
	}
)

const (
	defaultServer   = "https://kubernetes.default.svc"
	defaultProject  = "default"
	ArgoCDNamespace = "argocd"
)

func Build(argoNS, repoURL string, m *config.Manifest) (res.Resources, error) {
	// Without a RepositoryURL we can't do anything.
	if repoURL == "" {
		return res.Resources{}, nil
	}
	argoEnv, err := m.GetArgoironment()
	// If there's no ArgoCD environment, then we don't need to do anything.
	// if err != nil {
	// 	return res.Resources{}, nil
	// }
	if argoEnv == nil {
		return res.Resources{}, nil
	}

	files := make(res.Resources)
	eb := &argocdBuilder{repoURL: repoURL, files: files, argoEnv: argoEnv, argoNS: argoNS}
	err = m.Walk(eb)
	return eb.files, err
}

type argocdBuilder struct {
	repoURL string
	argoEnv *config.Argo
	files   res.Resources
	argoNS  string
}

func (b *argocdBuilder) Application(env *config.Environment, app *config.Application) error {
	// basePath := filepath.Join(config.PathForArgoEnvironment(b.argoEnv), "config")
	basePath := filepath.Join(config.PathForArgoEnvironment(), "config")
	argoFiles := res.Resources{}
	filename := filepath.Join(basePath, env.Name+"-"+app.Name+"-app.yaml")
	argoFiles[filename] = makeApplication(env.Name+"-"+app.Name, b.argoNS, defaultProject, env.Name, defaultServer, makeSource(env, app, b.repoURL))
	b.files = res.Merge(argoFiles, b.files)
	err := argoEnvironmentResources(b.argoEnv, b.files)
	if err != nil {
		return err
	}
	return nil
}

func argoEnvironmentResources(env *config.Argo, files res.Resources) error {
	if env.Namespace == "" {
		return nil
	}
	// basePath := filepath.Join(config.PathForArgoEnvironment(env), "config")
	basePath := filepath.Join(config.PathForArgoEnvironment(), "config")
	filename := filepath.Join(basePath, "kustomization.yaml")
	resourceNames := []string{}
	for k, _ := range files {
		resourceNames = append(resourceNames, filepath.Base(k))
	}
	sort.Strings(resourceNames)
	files[filename] = &res.Kustomization{Resources: resourceNames}
	return nil
}

func makeSource(env *config.Environment, app *config.Application, repoURL string) argoappv1.ApplicationSource {
	if app.ConfigRepo == nil {
		return argoappv1.ApplicationSource{
			RepoURL: repoURL,
			Path:    filepath.Join(config.PathForApplication(env, app), "base"),
		}
	}
	return argoappv1.ApplicationSource{
		RepoURL:        app.ConfigRepo.URL,
		Path:           app.ConfigRepo.Path,
		TargetRevision: app.ConfigRepo.TargetRevision,
	}
}

func makeApplication(appName, argoNS, project, ns, server string, source argoappv1.ApplicationSource) *argoappv1.Application {
	return &argoappv1.Application{
		TypeMeta:   applicationTypeMeta,
		ObjectMeta: meta.ObjectMeta(meta.NamespacedName(argoNS, appName)),
		Spec: argoappv1.ApplicationSpec{
			Project: project,
			Destination: argoappv1.ApplicationDestination{
				Namespace: ns,
				Server:    server,
			},
			Source:     source,
			SyncPolicy: syncPolicy,
		},
	}
}
