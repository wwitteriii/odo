package environments

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/openshift/odo/pkg/pipelines/config"
	"github.com/openshift/odo/pkg/pipelines/meta"
	"github.com/openshift/odo/pkg/pipelines/namespaces"
	res "github.com/openshift/odo/pkg/pipelines/resources"
	"github.com/openshift/odo/pkg/pipelines/roles"
	"github.com/spf13/afero"
	v1 "k8s.io/api/rbac/v1"
)

const kustomization = "kustomization.yaml"

type envBuilder struct {
	files           res.Resources
	pipelinesConfig *config.PipelinesConfig
	fs              afero.Fs
	saName          string
	manifest        *config.Manifest
}

func Build(fs afero.Fs, m *config.Manifest, saName string) (res.Resources, error) {
	files := res.Resources{}
	cfg := m.GetPipelinesConfig()
	eb := &envBuilder{fs: fs, files: files, pipelinesConfig: cfg, saName: saName, manifest: m}
	return eb.files, m.Walk(eb)
}

func (b *envBuilder) Application(app *config.Application, env *config.Environment) error {
	appPath := filepath.Join(config.PathForApplication(env, app))
	appFiles, err := filesForApplication(env, appPath, app, b.manifest)
	if err != nil {
		return err
	}
	b.files = res.Merge(appFiles, b.files)
	return nil
}

func (b *envBuilder) Service(env *config.Environment, svc *config.Service) error {
	svcPath := config.PathForService(env, svc.Name)
	svcFiles, err := filesForService(svcPath, svc)
	if err != nil {
		return err
	}
	b.files = res.Merge(svcFiles, b.files)
	// RoleBinding is created only when an environment has a service and the
	// CICD environment is defined.
	if b.pipelinesConfig == nil {
		return nil
	}
	envBasePath := filepath.Join(config.PathForEnvironment(env), "env", "base")
	envBindingPath := filepath.Join(envBasePath, fmt.Sprintf("%s-rolebinding.yaml", env.Name))
	if _, ok := b.files[envBindingPath]; !ok {
		b.files[envBindingPath] = createRoleBinding(env, envBasePath, b.pipelinesConfig.Name, b.saName)
	}
	return nil
}

func (b *envBuilder) Environment(env *config.Environment) error {
	envPath := filepath.Join(config.PathForEnvironment(env), "env")
	basePath := filepath.Join(envPath, "base")
	envFiles := filesForEnvironment(basePath, env)
	kustomizedFilenames, err := ListFiles(b.fs, basePath)
	if err != nil {
		return fmt.Errorf("failed to list initial files for %s: %s", basePath, err)
	}
	envBindingPath := filepath.Join(basePath, fmt.Sprintf("%s-rolebinding.yaml", env.Name))
	if _, ok := b.files[envBindingPath]; ok {
		envFiles[envBindingPath] = b.files[envBindingPath]
	}

	for k := range envFiles {
		kustomizedFilenames[filepath.Base(k)] = true
	}
	envFiles[filepath.Join(basePath, kustomization)] = &res.Kustomization{Resources: ExtractFilenames(kustomizedFilenames)}
	overlaysPath := filepath.Join(envPath, "overlays")
	relPath, err := filepath.Rel(overlaysPath, basePath)
	if err != nil {
		return err
	}
	envFiles[filepath.Join(overlaysPath, kustomization)] = &res.Kustomization{Bases: []string{relPath}}
	b.files = res.Merge(envFiles, b.files)
	return nil
}

func filesForEnvironment(basePath string, env *config.Environment) res.Resources {
	envFiles := res.Resources{}
	filename := filepath.Join(basePath, fmt.Sprintf("%s-environment.yaml", env.Name))
	envFiles[filename] = namespaces.Create(env.Name)
	return envFiles
}

func filesForApplication(env *config.Environment, appPath string, app *config.Application, manifest *config.Manifest) (res.Resources, error) {
	envFiles := res.Resources{}
	basePath := filepath.Join(appPath, "base")
	overlaysPath := filepath.Join(appPath, "overlays")
	overlaysFile := filepath.Join(overlaysPath, kustomization)
	overlayRel, err := filepath.Rel(overlaysPath, basePath)
	if err != nil {
		return nil, err
	}
	baseKustomization := filepath.Join(appPath, "base", kustomization)
	relServices := []string{}

	for _, environment := range app.Environments {
		currentEnv := manifest.GetEnvironment(environment.Ref)
		for _, serRef := range environment.ServiceRefs {
			svcPath := config.PathForService(currentEnv, serRef)
			relService, err := filepath.Rel(filepath.Dir(baseKustomization), svcPath)
			if err != nil {
				return nil, err
			}
			relServices = append(relServices, relService)
		}
	}
	fmt.Println("This is the app path", appPath)
	envFiles[filepath.Join(appPath, kustomization)] = &res.Kustomization{Bases: []string{"overlays"}}
	envFiles[filepath.Join(appPath, "base", kustomization)] = &res.Kustomization{Bases: relServices}
	envFiles[overlaysFile] = &res.Kustomization{Bases: []string{overlayRel}}

	return envFiles, nil
}

func createRoleBinding(env *config.Environment, basePath, cicdNS, saName string) *v1.RoleBinding {
	sa := roles.CreateServiceAccount(meta.NamespacedName(cicdNS, saName))
	return roles.CreateRoleBinding(meta.NamespacedName(env.Name, fmt.Sprintf("%s-rolebinding", env.Name)), sa, "ClusterRole", "edit")
}

func filesForService(svcPath string, app *config.Service) (res.Resources, error) {
	envFiles := res.Resources{}
	basePath := filepath.Join(svcPath, "base")
	overlaysPath := filepath.Join(svcPath, "overlays")
	overlaysFile := filepath.Join(overlaysPath, kustomization)
	overlayRel, err := filepath.Rel(overlaysPath, basePath)
	if err != nil {
		return nil, err
	}
	envFiles[filepath.Join(svcPath, kustomization)] = &res.Kustomization{Bases: []string{"overlays"}}
	envFiles[filepath.Join(svcPath, "base", kustomization)] = &res.Kustomization{Bases: []string{"./config"}}
	envFiles[overlaysFile] = &res.Kustomization{Bases: []string{overlayRel}}

	return envFiles, nil
}

func ExtractFilenames(f map[string]bool) []string {
	names := []string{}
	for k := range f {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}

func ListFiles(fs afero.Fs, base string) (map[string]bool, error) {
	files := map[string]bool{}
	err := afero.Walk(fs, base, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if info.IsDir() {
			return nil
		}
		filename := strings.TrimPrefix(path, base+"/")
		if filename == kustomization {
			return nil
		}
		files[filename] = true
		return nil
	})
	return files, err
}
