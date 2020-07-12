package scm

import (
	"net/url"
	"strings"

	triggersv1 "github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
)

const (
	gitlabPushEventFilters = "header.match('X-Gitlab-Event','Push Hook') && body.project.path_with_namespace == '%s'"
	gitlabType             = "gitlab"
)

type gitlabSpec struct {
	pushBinding string
}

func init() {
	gits[gitlabType] = newGitLab
}

func newGitLab(rawURL string) (Repository, error) {
	path, err := processRawURL(rawURL, proccessGitLabPath)
	if err != nil {
		return nil, err
	}
	return &repository{url: rawURL, path: path, spec: &gitlabSpec{pushBinding: "gitlab-push-binding"}}, nil
}

func proccessGitLabPath(parsedURL *url.URL) (string, error) {
	components, err := splitRepositoryPath(parsedURL)
	if err != nil {
		return "", err
	}
	if len(components) < 2 {
		return "", invalidRepoPathError(gitlabType, parsedURL.Path)
	}
	path := strings.Join(components, "/")
	return path, nil
}

func (r *gitlabSpec) pushBindingName() string {
	return r.pushBinding
}

func (r *gitlabSpec) pushBindingParams() []triggersv1.Param {
	return []triggersv1.Param{
		createBindingParam("io.openshift.build.commit.ref", "$(body.ref)"),
		createBindingParam("io.openshift.build.commit.id", "$(body.after)"),
		createBindingParam("gitrepositoryurl", "$(body.project.git_http_url)"),
		createBindingParam("fullname", "$(body.project.path_with_namespace)"),
		// createBindingParam("io.openshift.build.commit.date", "$(body.head_commit.timestamp)"),
		// createBindingParam("io.openshift.build.commit.message", "$(body.head_commit.message"),
		createBindingParam("io.openshift.build.commit.author", "$(body.user_username)"),
	}
}

func (r *gitlabSpec) pushEventFilters() string {
	return gitlabPushEventFilters
}

func (r *gitlabSpec) eventInterceptor(secretNamespace, secretName string) *triggersv1.EventInterceptor {
	return &triggersv1.EventInterceptor{
		GitLab: &triggersv1.GitLabInterceptor{
			SecretRef: &triggersv1.SecretRef{
				SecretName: secretName,
				SecretKey:  webhookSecretKey,
				Namespace:  secretNamespace,
			},
		},
	}
}
