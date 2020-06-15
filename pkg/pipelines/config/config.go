package config

import (
	"fmt"
	"path/filepath"
	"sort"
)

// PathForService gives a repo-rooted path within a repository.
func PathForService(env *Environment, serviceName string) string {
	return filepath.Join(PathForEnvironment(env), "services", serviceName)
}

// PathForApplication generates a repo-rooted path within a repository.
func PathForApplication(env *Environment, app *Application) string {
	return filepath.Join(PathForEnvironment(env), "apps", app.Name)
}

// PathForEnvironment gives a repo-rooted path within a repository.
func PathForEnvironment(env *Environment) string {
	return filepath.Join("environments", env.Name)
}

// PathForPipelines returns the path only for the CICD environment.
func PathForPipelines(pipeline *PipelinesConfig) string {
	return filepath.Join("config", pipeline.Name)
}

// PathForArgoCD returns the path for recording ArgoCD configuration.
func PathForArgoCD() string {
	return filepath.Join("config", "argocd")
}

// Manifest describes a set of environments, apps and services for deployment.
type Manifest struct {
	GitOpsURL    string         `json:"gitops_url,omitempty"`
	Environments []*Environment `json:"environments,omitempty"`
	Config       *Config        `json:"config,omitempty"`
	Apps         []*Application `json:"apps,omitempty"`
}

// GetEnvironment returns a named environment if it exists in the configuration.
func (m *Manifest) GetEnvironment(n string) *Environment {
	for _, env := range m.Environments {
		if env.Name == n {
			return env
		}
	}
	return nil
}

// GetApplication returns a named application, within an environment, if it
// exists.
// func (m *Manifest) GetApplication(environment, application string) *Application {
// 	for _, env := range m.Environments {
// 		if env.Name == environment {
// 			for _, app := range m.Apps {
// 				if app.Name == application {
// 					return app
// 				}
// 			}
// 		}
// 	}
// 	return nil
// }

func (m *Manifest) GetApplication(reqAppName string) *Application {
	for _, App := range m.Apps {
		if App.Name == reqAppName {
			return App
		}
	}
	return nil
}

// AddService adds a new service to a specific environment and creates a
// reference to it within an Application.
func (m *Manifest) AddService(envName, appName string, svc *Service) error {
	fmt.Println("The Addservice field was enetered twice")
	env := m.GetEnvironment(envName)
	if env == nil {
		return fmt.Errorf("environment %s does not exist", envName)
	}
	for _, service := range env.Services {
		if service.Name == svc.Name {
			return fmt.Errorf("service %s already exists at %s", svc.Name, env.Name)
		}
	}
	app := m.GetApplication(appName)
	if app == nil {
		app = &Application{
			Name: appName,
			Environments: []*EnvironmentRefs{
				{
					Ref: envName,
				},
			},
		}
	}
	env.Services = append(env.Services, svc)
	app = appendServiceRef(app, envName, svc.Name)
	m.Apps = append(m.Apps, app)

	// app.ServiceRefs = append(app.ServiceRefs, svc.Name)
	return nil
}

func appendServiceRef(app *Application, envName, svcName string) *Application {
	for _, env := range app.Environments {
		if env.Ref == envName {
			env.ServiceRefs = append(env.ServiceRefs, svcName)
		}
		return app
	}
	return nil
}

// GetPipelinesConfig returns the global Pipelines configuration, if one exists.
func (m *Manifest) GetPipelinesConfig() *PipelinesConfig {
	if m.Config != nil {
		return m.Config.Pipelines
	}
	return nil
}

// GetArgoCDConfig returns the global ArgoCD configuration, if one exists.
func (m *Manifest) GetArgoCDConfig() *ArgoCDConfig {
	if m.Config != nil {
		return m.Config.ArgoCD
	}
	return nil

}

func (m *Manifest) GetService(envName, svcName string) *Service {
	env := m.GetEnvironment(envName)
	if env != nil {
		for _, svc := range env.Services {
			if svc.Name == svcName {
				return svc
			}
		}
	}
	return nil
}

// Environment is a slice of Apps, these are the named apps in the namespace.
//
type Environment struct {
	Name      string     `json:"name,omitempty"`
	Pipelines *Pipelines `json:"pipelines,omitempty"`
	Services  []*Service `json:"services,omitempty"`
}

// Config represents the configuration for non-application environments.
type Config struct {
	Pipelines *PipelinesConfig `json:"pipelines,omitempty"`
	ArgoCD    *ArgoCDConfig    `json:"argocd,omitempty"`
}

// PipelinesConfig provides configuration for the CI/CD pipelines.
type PipelinesConfig struct {
	Name string `json:"name,omitempty"`
}

// ArgoCDConfig provides configuration for the ArgoCD application generation.
type ArgoCDConfig struct {
	Namespace string `json:"namespace,omitempty"`
}

// GoString return environment name
// func (e Environment) GoString() string {
// 	return e.Name
// }

// Application has many services.
//
// The ConfigRepo indicates that the configuration for this application lives in
// another repository.
type Application struct {
	Name         string             `json:"name,omitempty"`
	Environments []*EnvironmentRefs `json:"environments,omitempty"`
	ConfigRepo   *Repository        `json:"config_repo,omitempty"`
}

//EnvironmentRefs are environment references
type EnvironmentRefs struct {
	Ref         string   `json:"ref,omitempty"`
	ServiceRefs []string `json:"serviceRefs,omitempty"`
}

// Service has an upstream source.
type Service struct {
	Name      string     `json:"name,omitempty"`
	Webhook   *Webhook   `json:"webhook,omitempty"`
	SourceURL string     `json:"source_url,omitempty"`
	Pipelines *Pipelines `json:"pipelines,omitempty"`
}

// Webhook provides Github webhook secret for eventlisteners
type Webhook struct {
	Secret *Secret `json:"secret,omitempty"`
}

// Secret represents a K8s secret in a namespace
type Secret struct {
	Name      string `json:"name,omitempty"`
	Namespace string `json:"namespace,omitempty"`
}

// Repository refers to an upstream source for reading additional config from.
type Repository struct {
	URL string `json:"url,omitempty"`
	// TargetRevision defines the commit, tag, or branch in which to sync the application to.
	// If omitted, will sync to HEAD
	TargetRevision string `json:"target_revision,omitempty"`
	// Path is a directory path within the Git repository.
	Path string `json:"path,omitempty"`
}

// Pipelines describes the names for pipelines to be executed for CI and CD.
//
// These pipelines will be executed with a Git clone URL and commit SHA.
type Pipelines struct {
	Integration *TemplateBinding `json:"integration,omitempty"`
}

// TemplateBinding is a combination of the template and binding to be used for a
// pipeline execution.
type TemplateBinding struct {
	Template string   `json:"template,omitempty"`
	Bindings []string `json:"bindings,omitempty"`
}

// Walk implements post-node visiting of each element in the manifest.
//
// Every App, Service and Environment is called once, and any error from the
// handling function terminates the Walk.
//
// The environments are sorted using a custom sorting mechanism, that orders by
// name, but, moves CICD environments to the bottom of the list.
func (m Manifest) Walk(visitor interface{}) error {
	sort.Sort(ByName(m.Environments))
	for _, env := range m.Environments {
		for _, svc := range env.Services {
			if v, ok := visitor.(ServiceVisitor); ok {
				err := v.Service(env, svc)
				if err != nil {
					return err
				}
			}
		}

		if m.GetEnvironment(env.Name) != nil {
			if v, ok := visitor.(EnvironmentVisitor); ok {
				err := v.Environment(env)
				if err != nil {
					return err
				}
			}
		}
	}

	for _, app := range m.Apps {
		if v, ok := visitor.(ApplicationVisitor); ok {
			for _, env := range app.Environments {
				err := v.Application(app, m.GetEnvironment(env.Ref))
				if err != nil {
					return err
				}
			}

		}

	}

	return nil
}

type ByName []*Environment

func (a ByName) Len() int      { return len(a) }
func (a ByName) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByName) Less(i, j int) bool {
	return a[i].Name < a[j].Name
}
