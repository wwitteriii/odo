package scm

import (
	"net/url"

	"github.com/openshift/odo/pkg/pipelines/meta"
	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	triggersv1 "github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
)

const (
	gitlabPRBindingName   = "gitlab-pr-binding"
	gitlabPushBindingName = "gitlab-push-binding"
)

type GitlabRepository struct {
	URL *url.URL
}

func NewGitlabRepository(rawURL string) (*GitlabRepository, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	return &GitlabRepository{URL: parsedURL}, nil
}

func (repo *GitlabRepository) CreateGitlabPRBinding(ns string) (triggersv1.TriggerBinding, string) {
	return triggersv1.TriggerBinding{
		TypeMeta:   triggerBindingTypeMeta,
		ObjectMeta: meta.ObjectMeta(meta.NamespacedName(ns, gitlabPRBindingName)),
		Spec: triggersv1.TriggerBindingSpec{
			Params: []pipelinev1.Param{
				createBindingParam("gitref", "$(body.object_attributes.source_branch)"),
				createBindingParam("gitsha", "$(body.object_attributes.last_commit.id)"),
				createBindingParam("gitrepositoryurl", "$(body.project.git_http_url)"),
				createBindingParam("fullname", "$(body.project.path_with_namespace)"),
			},
		},
	}, gitlabPRBindingName

}

func (repo *GitlabRepository) CreateGitlabPushBinding(ns string) (triggersv1.TriggerBinding, string) {
	return triggersv1.TriggerBinding{
		TypeMeta:   triggerBindingTypeMeta,
		ObjectMeta: meta.ObjectMeta(meta.NamespacedName(ns, gitlabPushBindingName)),
		Spec: triggersv1.TriggerBindingSpec{
			Params: []pipelinev1.Param{
				createBindingParam("gitref", "$(body.object_attributes.source_branch)"),
				createBindingParam("gitsha", "$(body.object_attributes.last_commit.id)"),
				createBindingParam("gitrepositoryurl", "$(body.project.git_http_url)"),
			},
		},
	}, gitlabPushBindingName
}
