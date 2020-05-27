package scm

import (
	"net/url"
	"strings"

	"github.com/openshift/odo/pkg/pipelines/meta"
	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	"github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
	triggersv1 "github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
)

const (
	gitlabPRBindingName   = "gitlab-pr-binding"
	gitlabPushBindingName = "gitlab-push-binding"

	gitlabCIDryRunFilters = "( header.match ( ‘X-Gitlab-Event’ , ‘Merge Request Hook’ ) && body.object_kind == ‘merge_request’ ) && body.object_attributes.state == ‘opened’ && body.project.path_with_namespace == %s  && body.project.default_branch == body.object_attributes.target_branch )"
	gitlabCDDeployFilters = "( header.match ( ‘X-Gitlab-Event’ , ‘Push Hook’ ) && body.object_kind == ‘push’ && body.project.path_with_namespace == %s && body.ref.EndsWith (body.project.default_branch )"
)

type GitlabRepository struct {
	URL  *url.URL
	path string // GitLab repo path eg: (org/name)
}

func NewGitlabRepository(rawURL string) (*GitlabRepository, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	var components []string
	for _, s := range strings.Split(parsedURL.Path, "/") {
		if s != "" {
			components = append(components, s)
		}
	}
	if len(components) < 2 {
		return nil, invalidRepoPathError(rawURL)
	}
	path := components[0] + "/" + strings.TrimSuffix(components[1], ".git")
	return &GitlabRepository{URL: parsedURL, path: path}, nil

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

func (repo *GitlabRepository) CreateCITrigger(name, secretName, secretNS, template string, bindings []string) v1alpha1.EventListenerTrigger {
	return triggersv1.EventListenerTrigger{
		Name: name,
		Interceptors: []*triggersv1.EventInterceptor{
			createEventInterceptor(gitlabCIDryRunFilters, repo.path),
			repo.CreateInterceptor(secretName, secretNS),
		},
		Bindings: createBindings(bindings),
		Template: createListenerTemplate(template),
	}
}

func (repo *GitlabRepository) CreateCDTrigger(name, secretName, secretNS, template string, bindings []string) v1alpha1.EventListenerTrigger {
	return triggersv1.EventListenerTrigger{
		Name: name,
		Interceptors: []*triggersv1.EventInterceptor{
			createEventInterceptor(gitlabCDDeployFilters, repo.path),
			repo.CreateInterceptor(secretName, secretNS),
		},
		Bindings: createBindings(bindings),
		Template: createListenerTemplate(template),
	}
}

func (repo *GitlabRepository) CreateInterceptor(secretName, secretNs string) *triggersv1.EventInterceptor {
	return &triggersv1.EventInterceptor{
		GitLab: &triggersv1.GitLabInterceptor{
			SecretRef: &triggersv1.SecretRef{
				SecretName: secretName,
				SecretKey:  webhookSecretKey,
				Namespace:  secretNs,
			},
		},
	}
}
