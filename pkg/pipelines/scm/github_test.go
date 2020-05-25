package scm

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/h2non/gock"
	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	triggersv1 "github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ Repository = (*GitHubRepository)(nil)

var mockHeaders = map[string]string{
	"X-GitHub-Request-Id":   "DD0E:6011:12F21A8:1926790:5A2064E2",
	"X-RateLimit-Limit":     "60",
	"X-RateLimit-Remaining": "59",
	"X-RateLimit-Reset":     "1512076018",
}

func TestCreatePRBindingForGithub(t *testing.T) {
	repo, err := NewGitHubRepository("http://github.com/org/test", "")
	assertNoError(t, err)
	want := triggersv1.TriggerBinding{
		TypeMeta: triggerBindingTypeMeta,
		ObjectMeta: v1.ObjectMeta{
			Name:      "github-pr-binding",
			Namespace: "testns",
		},
		Spec: triggersv1.TriggerBindingSpec{
			Params: []pipelinev1.Param{
				{
					Name: "gitref",
					Value: pipelinev1.ArrayOrString{
						StringVal: "$(body.pull_request.head.ref)",
						Type:      pipelinev1.ParamTypeString,
					},
				},
				{
					Name: "gitsha",
					Value: pipelinev1.ArrayOrString{
						StringVal: "$(body.pull_request.head.sha)",
						Type:      pipelinev1.ParamTypeString,
					},
				},
				{
					Name: "gitrepositoryurl",
					Value: pipelinev1.ArrayOrString{
						StringVal: "$(body.repository.clone_url)",
						Type:      pipelinev1.ParamTypeString,
					},
				},
				{
					Name: "fullname",
					Value: pipelinev1.ArrayOrString{
						StringVal: "$(body.repository.full_name)",
						Type:      pipelinev1.ParamTypeString,
					},
				},
			},
		},
	}
	got, name := repo.CreatePRBinding("testns")
	if name != githubPRBindingName {
		t.Fatalf("CreatePushBinding() returned a wrong binding: want %v got %v", githubPRBindingName, name)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("createPRBinding() failed:\n%s", diff)
	}
}

func TestCreatePushBindingForGithub(t *testing.T) {
	repo, err := NewGitHubRepository("http://github.com/org/test", "")
	assertNoError(t, err)
	want := triggersv1.TriggerBinding{
		TypeMeta: triggerBindingTypeMeta,
		ObjectMeta: v1.ObjectMeta{
			Name:      "github-push-binding",
			Namespace: "testns",
		},
		Spec: triggersv1.TriggerBindingSpec{
			Params: []pipelinev1.Param{
				{
					Name: "gitref",
					Value: pipelinev1.ArrayOrString{
						StringVal: "$(body.ref)",
						Type:      pipelinev1.ParamTypeString,
					},
				},
				{
					Name: "gitsha",
					Value: pipelinev1.ArrayOrString{
						StringVal: "$(body.head_commit.id)",
						Type:      pipelinev1.ParamTypeString,
					},
				},
				{
					Name: "gitrepositoryurl",
					Value: pipelinev1.ArrayOrString{
						StringVal: "$(body.repository.clone_url)",
						Type:      pipelinev1.ParamTypeString,
					},
				},
			},
		},
	}
	got, name := repo.CreatePushBinding("testns")
	if name != githubPushBindingName {
		t.Fatalf("CreatePushBinding() returned a wrong binding: want %v got %v", githubPushBindingName, name)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("CreatePushBinding() failed:\n%s", diff)
	}
}

func TestCreateCITriggerForGithub(t *testing.T) {
	repo, err := NewGitHubRepository("http://github.com/org/test", "")
	assertNoError(t, err)
	want := triggersv1.EventListenerTrigger{
		Name: "test",
		Bindings: []*triggersv1.EventListenerBinding{
			&triggersv1.EventListenerBinding{Name: "test-binding"},
		},
		Template: triggersv1.EventListenerTemplate{Name: "test-template"},
		Interceptors: []*triggersv1.EventInterceptor{
			&triggersv1.EventInterceptor{
				CEL: &triggersv1.CELInterceptor{
					Filter: fmt.Sprintf(githubCIDryRunFilters, "org/test"),
				},
			},
			&triggersv1.EventInterceptor{
				GitHub: &triggersv1.GitHubInterceptor{
					SecretRef: &triggersv1.SecretRef{SecretKey: "webhook-secret-key", SecretName: "secret", Namespace: "ns"},
				},
			},
		},
	}
	got := repo.CreateCITrigger("test", "secret", "ns", "test-template", []string{"test-binding"})
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("CreateCITrigger() failed:\n%s", diff)
	}
}

func TestCreateCDTriggersForGithub(t *testing.T) {
	repo, err := NewGitHubRepository("http://github.com/org/test", "")
	assertNoError(t, err)
	want := triggersv1.EventListenerTrigger{
		Name: "test",
		Bindings: []*triggersv1.EventListenerBinding{
			&triggersv1.EventListenerBinding{Name: "test-binding"},
		},
		Template: triggersv1.EventListenerTemplate{Name: "test-template"},
		Interceptors: []*triggersv1.EventInterceptor{
			&triggersv1.EventInterceptor{
				CEL: &triggersv1.CELInterceptor{
					Filter: fmt.Sprintf(githubCDDeployFilters, "org/test"),
				},
			},
			&triggersv1.EventInterceptor{
				GitHub: &triggersv1.GitHubInterceptor{
					SecretRef: &triggersv1.SecretRef{SecretKey: "webhook-secret-key", SecretName: "secret", Namespace: "ns"},
				},
			},
		},
	}
	got := repo.CreateCDTrigger("test", "secret", "ns", "test-template", []string{"test-binding"})
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("CreateCDTrigger() failed:\n%s", diff)
	}
}

func TestNewGitHubRepository(t *testing.T) {
	tests := []struct {
		url      string
		repoPath string
		errMsg   string
	}{
		{
			"http://github.org",
			"",
			"unable to determine repo path from: http://github.org",
		},
		{
			"http://github.com/",
			"",
			"unable to determine repo path from: http://github.com/",
		},
		{
			"http://github.com/foo/bar",
			"foo/bar",
			"",
		},
		{
			"https://githuB.com/foo/bar.git",
			"foo/bar",
			"",
		},
		{
			"https://githuB.com/foo/bar/test.git",
			"foo/bar",
			"",
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("Test %d", i), func(rt *testing.T) {
			repo, err := NewGitHubRepository(tt.url, "")
			if err != nil {
				if diff := cmp.Diff(tt.errMsg, err.Error()); diff != "" {
					rt.Fatalf("repo path errMsg mismatch: \n%s", diff)
				}
			}
			if repo != nil {
				if diff := cmp.Diff(tt.repoPath, repo.path); diff != "" {
					rt.Fatalf("repo path mismatch: got\n%s", diff)
				}
			}
		})
	}
}

func TestWebhookWithFakeClient(t *testing.T) {

	repo, err := NewGitHubRepository("https://fake.com/foo/bar.git", "token")
	if err != nil {
		t.Fatal(err)
	}

	listenerURL := "http://example.com/webhook"
	ids, err := repo.ListWebhooks(listenerURL)
	assertNoError(t, err)

	// start with no webhooks
	if len(ids) > 0 {
		t.Fatal(err)
	}

	// create a webhook
	id, err := repo.CreateWebhook(listenerURL, "secret")
	assertNoError(t, err)
	if len(ids) > 0 {
		t.Fatal(err)
	}

	// verify and remember our ID
	if id == "" {
		t.Fatal(err)
	}

	// list again
	ids, err = repo.ListWebhooks(listenerURL)
	assertNoError(t, err)

	// verify ID from list
	if diff := cmp.Diff(ids, []string{id}); diff != "" {
		t.Fatalf("created id mismatch got\n%s", diff)
	}

	// delete webhook
	deleted, err := repo.DeleteWebhooks(ids)
	assertNoError(t, err)

	// verify deleted IDs
	if diff := cmp.Diff(ids, deleted); diff != "" {
		t.Fatalf("deleted ids mismatch got\n%s", diff)
	}

	ids, err = repo.ListWebhooks(listenerURL)
	assertNoError(t, err)

	// verify no webhooks
	if len(ids) > 0 {
		t.Fatal(err)
	}
}

func TestListWebHooks(t *testing.T) {

	defer gock.Off()

	gock.New("https://api.github.com").
		Get("/repos/foo/bar/hooks").
		Reply(200).
		Type("application/json").
		SetHeaders(mockHeaders).
		File("testdata/hooks.json")

	repo, err := NewGitHubRepository("https://github.com/foo/bar.git", "token")
	assertNoError(t, err)

	ids, err := repo.ListWebhooks("http://example.com/webhook")
	assertNoError(t, err)

	if diff := cmp.Diff(ids, []string{"1"}); diff != "" {
		t.Errorf("driver errMsg mismatch got\n%s", diff)
	}
}

func TestDeleteWebHooks(t *testing.T) {

	defer gock.Off()

	gock.New("https://api.github.com").
		Delete("/repos/foo/bar/hooks/1").
		Reply(204).
		Type("application/json").
		SetHeaders(mockHeaders)

	repo, err := NewGitHubRepository("https://github.com/foo/bar.git", "token")
	assertNoError(t, err)

	deleted, err := repo.DeleteWebhooks([]string{"1"})
	assertNoError(t, err)

	if diff := cmp.Diff([]string{"1"}, deleted); diff != "" {
		t.Errorf("deleted mismatch got\n%s", diff)
	}
}

func TestCreateWebHook(t *testing.T) {

	defer gock.Off()

	gock.New("https://api.github.com").
		Post("/repos/foo/bar/hooks").
		Reply(201).
		Type("application/json").
		SetHeaders(mockHeaders).
		File("testdata/hook.json")

	repo, err := NewGitHubRepository("https://github.com/foo/bar.git", "token")
	assertNoError(t, err)

	created, err := repo.CreateWebhook("http://example.com/webhook", "mysecret")
	assertNoError(t, err)

	if diff := cmp.Diff("1", created); diff != "" {
		t.Errorf("deleted mismatch got\n%s", diff)
	}
}
