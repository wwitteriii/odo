package pipelines

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/openshift/odo/pkg/pipelines/config"
	"github.com/openshift/odo/pkg/pipelines/ioutils"
	"github.com/openshift/odo/pkg/pipelines/meta"
	res "github.com/openshift/odo/pkg/pipelines/resources"
	"github.com/openshift/odo/pkg/pipelines/scm"
)

var testpipelineConfig = &config.PipelinesConfig{Name: "tst-cicd"}
var testArgoCDConfig = &config.ArgoCDConfig{Namespace: "tst-argocd"}
var Config = &config.Config{ArgoCD: testArgoCDConfig, Pipelines: testpipelineConfig}

func TestCreateManifest(t *testing.T) {
	repoURL := "https://github.com/foo/bar.git"
	want := &config.Manifest{
		GitOpsURL: repoURL,
		Config:    Config,
	}
	got := createManifest(repoURL, Config)
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("pipelines didn't match: %s\n", diff)
	}
}

func TestInitialFiles(t *testing.T) {
	prefix := "tst-"
	gitOpsURL := "https://github.com/foo/test-repo"
	gitOpsWebhook := "123"
	o := InitOptions{Prefix: prefix, GitOpsWebhookSecret: gitOpsWebhook, DockerConfigJSONFilename: "", SealedSecretsService: meta.NamespacedName("", "")}

	defer stubDefaultPublicKeyFunc(t)()
	fakeFs := ioutils.NewMapFilesystem()
	repo, err := scm.NewRepository(gitOpsURL)
	assertNoError(t, err)
	got, err := createInitialFiles(fakeFs, repo, &o)
	assertNoError(t, err)

	want := res.Resources{
		pipelinesFile: createManifest(gitOpsURL, &config.Config{Pipelines: testpipelineConfig}),
	}
	resources, err := createCICDResources(fakeFs, repo, testpipelineConfig, &o)
	if err != nil {
		t.Fatalf("CreatePipelineResources() failed due to :%s\n", err)
	}
	files := getResourceFiles(resources)

	want = res.Merge(addPrefixToResources("config/tst-cicd/base", resources), want)
	want = res.Merge(addPrefixToResources("config/tst-cicd", getCICDKustomization(files)), want)

	if diff := cmp.Diff(want, got, cmpopts.IgnoreMapEntries(ignoreSecrets)); diff != "" {
		t.Fatalf("outputs didn't match: %s\n", diff)
	}
}

func ignoreSecrets(k string, v interface{}) bool {
	return k == "config/tst-cicd/base/03-secrets/gitops-webhook-secret.yaml"
}

func TestGetCICDKustomization(t *testing.T) {
	want := res.Resources{
		"overlays/kustomization.yaml": res.Kustomization{
			Bases: []string{"../base"},
		},
		"base/kustomization.yaml": res.Kustomization{
			Resources: []string{"resource1", "resource2"},
		},
	}
	got := getCICDKustomization([]string{"resource1", "resource2"})
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("getCICDKustomization was not correct: %s\n", diff)
	}

}

func TestAddPrefixToResources(t *testing.T) {
	files := map[string]interface{}{
		"base/kustomization.yaml": map[string]interface{}{
			"resources": []string{},
		},
		"overlays/kustomization.yaml": map[string]interface{}{
			"bases": []string{"../base"},
		},
	}

	want := map[string]interface{}{
		"test-prefix/base/kustomization.yaml": map[string]interface{}{
			"resources": []string{},
		},
		"test-prefix/overlays/kustomization.yaml": map[string]interface{}{
			"bases": []string{"../base"},
		},
	}
	if diff := cmp.Diff(want, addPrefixToResources("test-prefix", files)); diff != "" {
		t.Fatalf("addPrefixToResources failed, diff %s\n", diff)
	}
}
