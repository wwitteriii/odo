package scm

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	triggersv1 "github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestPRbindingForGitlab(t *testing.T) {
	repo := fakeGitlabRepository(t, "https://gitlab.com/dpalodka/firstproject")
	want := triggersv1.TriggerBinding{
		TypeMeta: triggerBindingTypeMeta,
		ObjectMeta: v1.ObjectMeta{
			Name:      "gitlab-pr-binding",
			Namespace: "testns",
		},
		Spec: triggersv1.TriggerBindingSpec{
			Params: []pipelinev1.Param{
				{
					Name: "gitref",
					Value: pipelinev1.ArrayOrString{
						StringVal: "$(body.object_attributes.source_branch)",
						Type:      pipelinev1.ParamTypeString,
					},
				},
				{
					Name: "gitsha",
					Value: pipelinev1.ArrayOrString{
						StringVal: "$(body.object_attributes.last_commit.id)",
						Type:      pipelinev1.ParamTypeString,
					},
				},
				{
					Name: "gitrepositoryurl",
					Value: pipelinev1.ArrayOrString{
						StringVal: "$(body.project.git_http_url)",
						Type:      pipelinev1.ParamTypeString,
					},
				},
				{
					Name: "fullname",
					Value: pipelinev1.ArrayOrString{
						StringVal: "$(body.project.path_with_namespace)",
						Type:      pipelinev1.ParamTypeString,
					},
				},
			},
		},
	}
	got, name := repo.CreateGitlabPRBinding("testns")
	if name != gitlabPRBindingName {
		t.Fatalf("CreatePushBinding() returned a wrong binding: want %v got %v", gitlabPRBindingName, name)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("createPRBinding() failed:\n%s", diff)
	}
}

func TestCreatePushBindingForGitlab(t *testing.T) {
	repo := fakeGitlabRepository(t, "https://gitlab.com/dpalodka/firstproject")
	want := triggersv1.TriggerBinding{
		TypeMeta: triggerBindingTypeMeta,
		ObjectMeta: v1.ObjectMeta{
			Name:      "gitlab-push-binding",
			Namespace: "testns",
		},
		Spec: triggersv1.TriggerBindingSpec{
			Params: []pipelinev1.Param{
				{
					Name: "gitref",
					Value: pipelinev1.ArrayOrString{
						StringVal: "$(body.object_attributes.source_branch)",
						Type:      pipelinev1.ParamTypeString,
					},
				},
				{
					Name: "gitsha",
					Value: pipelinev1.ArrayOrString{
						StringVal: "$(body.object_attributes.last_commit.id)",
						Type:      pipelinev1.ParamTypeString,
					},
				},
				{
					Name: "gitrepositoryurl",
					Value: pipelinev1.ArrayOrString{
						StringVal: "$(body.project.git_http_url)",
						Type:      pipelinev1.ParamTypeString,
					},
				},
			},
		},
	}
	got, name := repo.CreateGitlabPushBinding("testns")
	if name != gitlabPushBindingName {
		t.Fatalf("CreatePushBinding() returned a wrong binding: want %v got %v", gitlabPushBindingName, name)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("CreatePushBinding() failed:\n%s", diff)
	}
}

func fakeGitlabRepository(t *testing.T, rawURL string) *GitlabRepository {
	repo, err := NewGitlabRepository("https://gitlab.com/dpalodka/firstproject")
	if err != nil {
		t.Fatal(err)
	}
	return repo
}
