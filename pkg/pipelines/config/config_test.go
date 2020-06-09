package config

import (
	"fmt"
	"path/filepath"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestManifestWalk(t *testing.T) {
	m := &Manifest{
		GitOpsURL: "https://github.com/...",
		Config: &Config{
			Pipelines: &PipelinesConfig{
				Name: "test-pipelines",
			},
			ArgoCD: &ArgoCDConfig{
				Namespace: "test-argocd",
			},
		},
		Environments: []*Environment{
			{
				Name: "development",
				Pipelines: &Pipelines{
					Integration: &TemplateBinding{
						Template: "dev-ci-template",
						Bindings: []string{"dev-ci-binding"},
					},
				},
				Services: []*Service{
					{
						Name:      "service-http",
						SourceURL: "https://github.com/myproject/myservice.git",
					},
					{Name: "service-redis"},
				},
			},
			{
				Name: "staging",
			},
			{
				Name: "production",
				Services: []*Service{
					{Name: "service-http"},
					{Name: "service-metrics"},
				},
			},
		},
		Apps: []*Application{
			{
				Name: "app-1",
				Environments: []*EnvironmentRefs{
					{
						Ref: "development",
						ServiceRefs: []string{
							"service-http",
							"service-redis",
						},
					},
					{
						Ref: "production",
						ServiceRefs: []string{
							"service-http",
							"service-metrics",
						},
					},
				},
			},
		},
	}

	v := &testVisitor{paths: []string{}}
	err := m.Walk(v)
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(v.paths)

	want := []string{
		"envs/development",
		"envs/development/apps/app-1",
		"envs/development/services/service-http",
		"envs/development/services/service-redis",
		"envs/production",
		"envs/production/apps/app-1",
		"envs/production/services/service-http",
		"envs/production/services/service-metrics",
		"envs/staging",
	}

	if diff := cmp.Diff(want, v.paths); diff != "" {
		t.Fatalf("tree files: %s", diff)
	}
}

func TestGetPipelinesConfig(t *testing.T) {
	cfg := &Config{
		Pipelines: &PipelinesConfig{
			Name: "cicd",
		},
	}

	envTests := []struct {
		name     string
		manifest *Manifest
		want     *PipelinesConfig
	}{
		{
			name:     "manifest with configuration",
			manifest: &Manifest{Config: cfg},
			want:     cfg.Pipelines,
		},
		{
			name:     "manifest with no configuration",
			manifest: &Manifest{},
			want:     nil,
		},
	}

	for i, tt := range envTests {
		t.Run(fmt.Sprintf("test %d", i), func(rt *testing.T) {
			m := tt.manifest
			got := m.GetPipelinesConfig()
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("%s: configuration did not match:\n%s", tt.name, diff)
			}
		})
	}
}

func TestGetArgoCDConfig(t *testing.T) {
	cfg := &Config{
		ArgoCD: &ArgoCDConfig{
			Namespace: "argocd",
		},
	}

	envTests := []struct {
		name     string
		manifest *Manifest
		want     *ArgoCDConfig
	}{
		{
			name:     "manifest with configuration",
			manifest: &Manifest{Config: cfg},
			want:     cfg.ArgoCD,
		},
		{
			name:     "manifest with no configuration",
			manifest: &Manifest{},
			want:     nil,
		},
	}

	for i, tt := range envTests {
		t.Run(fmt.Sprintf("test %d", i), func(rt *testing.T) {
			m := tt.manifest
			got := m.GetArgoCDConfig()
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("%s: configuration did not match:\n%s", tt.name, diff)
			}
		})
	}
}

func TestGetEnvironment(t *testing.T) {
	m := &Manifest{Environments: makeEnvs([]testEnv{{name: "prod"}, {name: "testing"}})}
	env := m.GetEnvironment("prod")
	if env.Name != "prod" {
		t.Fatalf("got the wrong environment back: %#v", env)
	}

	unknown := m.GetEnvironment("unknown")
	if unknown != nil {
		t.Fatalf("found an unknown env: %#v", unknown)
	}
}

func makeEnvs(ns []testEnv) []*Environment {
	n := make([]*Environment, len(ns))
	for i, v := range ns {
		n[i] = &Environment{Name: v.name}
	}
	return n

}

type testEnv struct {
	name string
}

type testVisitor struct {
	pipelineServices []string
	paths            []string
}

func (v *testVisitor) Service(env *Environment, svc *Service) error {
	v.paths = append(v.paths, filepath.Join("envs", env.Name, "services", svc.Name))
	v.pipelineServices = append(v.pipelineServices, filepath.Join("cicd", env.Name, svc.Name))
	return nil
}

func (v *testVisitor) Application(app *Application) error {
	for _, env := range app.Environments {
		v.paths = append(v.paths, filepath.Join("envs", env.Ref, "apps", app.Name))
	}
	return nil
}

func (v *testVisitor) Environment(env *Environment) error {
	if env.Name == "cicd" {
		v.paths = append(v.paths, v.pipelineServices...)
	}
	v.paths = append(v.paths, filepath.Join("envs", env.Name))
	return nil
}
