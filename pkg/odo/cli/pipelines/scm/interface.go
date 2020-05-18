package scm

import (
	triggersv1 "github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
)

type Repository interface {
	CreatePRBinding(string) (triggersv1.TriggerBinding, string)
	CreatePushBinding(string) (triggersv1.TriggerBinding, string)
	GetURL() string
}
