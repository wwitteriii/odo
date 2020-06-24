package statustracker

import (
	"fmt"

	ssv1alpha1 "github.com/bitnami-labs/sealed-secrets/pkg/apis/sealed-secrets/v1alpha1"
	"github.com/openshift/odo/pkg/pipelines/deployment"
	"github.com/openshift/odo/pkg/pipelines/meta"
	res "github.com/openshift/odo/pkg/pipelines/resources"
	"github.com/openshift/odo/pkg/pipelines/roles"
	"github.com/openshift/odo/pkg/pipelines/secrets"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	operatorName         = "commit-status-tracker"
	containerImage       = "quay.io/redhat-developer/commit-status-tracker:v0.0.2"
	rolePath             = "02-rolebindings/commit-status-tracker-role.yaml"
	roleBindingPath      = "02-rolebindings/commit-status-tracker-rolebinding.yaml"
	serviceAccountPath   = "02-rolebindings/commit-status-tracker-service-role.yaml"
	secretPath           = "03-secrets/commit-status-tracker.yaml"
	deploymentPath       = "10-commit-status-tracker/operator.yaml"
	commitStatusAppLabel = "commit-status-tracker-operator"
)

type secretSealer = func(types.NamespacedName, string, string) (*ssv1alpha1.SealedSecret, error)

var defaultSecretSealer secretSealer = secrets.CreateSealedSecret

var (
	roleRules = []rbacv1.PolicyRule{
		{
			APIGroups: []string{""},
			Resources: []string{"pods", "services", "services/finalizers", "endpoints", "persistentvolumeclaims", "events", "configmaps", "secrets"},
			Verbs:     []string{"create", "delete", "get", "list", "patch", "update", "watch"},
		},
		{
			APIGroups: []string{"apps"},
			Resources: []string{"deployments", "daemonsets", "replicasets", "statefulsets"},
			Verbs:     []string{"create", "delete", "get", "list", "patch", "update", "watch"},
		},
		{
			APIGroups: []string{"monitoring.coreos.com"},
			Resources: []string{"servicemonitors"},
			Verbs:     []string{"get", "create"},
		},
		{
			APIGroups:     []string{"apps"},
			Resources:     []string{"deployments/finalizers"},
			ResourceNames: []string{operatorName},
			Verbs:         []string{"update"},
		},
		{
			APIGroups: []string{""},
			Resources: []string{"pods"},
			Verbs:     []string{"get"},
		},
		{
			APIGroups: []string{"apps"},
			Resources: []string{"replicasets", "deployments"},
			Verbs:     []string{"get"},
		},
		{
			APIGroups: []string{"tekton.dev"},
			Resources: []string{"pipelineruns"},
			Verbs:     []string{"get", "list", "watch"},
		},
	}

	statusTrackerEnv = []corev1.EnvVar{
		{
			Name: "WATCH_NAMESPACE",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "metadata.namespace",
				},
			},
		},
		{
			Name: "POD_NAME",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "metadata.name",
				},
			},
		},
		{
			Name:  "OPERATOR_NAME",
			Value: operatorName,
		},
	}
)

func createStatusTrackerDeployment(ns string) *appsv1.Deployment {
	return deployment.Create(commitStatusAppLabel, ns, operatorName, containerImage,
		deployment.ServiceAccount(operatorName),
		deployment.Env(statusTrackerEnv),
		deployment.Command([]string{operatorName}))
}

// Resources returns a list of newly created resources that are required start
// the status-tracker service.
func Resources(ns, token string) (res.Resources, error) {
	name := meta.NamespacedName(ns, operatorName)
	sa := roles.CreateServiceAccount(name)

	secret, err := defaultSecretSealer(meta.NamespacedName(ns, "commit-status-tracker-git-secret"), token, "token")
	if err != nil {
		return nil, fmt.Errorf("failed to generate Status Tracker Secret: %v", err)
	}
	pipelineSA := roles.CreateServiceAccount(meta.NamespacedName(ns, "pipeline"))
	return res.Resources{
		serviceAccountPath: sa,
		secretPath:         secret,
		rolePath:           roles.CreateRole(name, roleRules),
		roleBindingPath:    roles.CreateRoleBindingForSubjects(name, "Role", operatorName, []rbacv1.Subject{{Kind: sa.Kind, Name: sa.Name, Namespace: sa.Namespace}, {Kind: pipelineSA.Kind, Name: pipelineSA.Name, Namespace: pipelineSA.Namespace}}),
		deploymentPath:     createStatusTrackerDeployment(ns),
	}, nil
}

func ptr32(i int32) *int32 {
	return &i
}
