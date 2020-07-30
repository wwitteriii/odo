package ui

import (
	"fmt"

	"gopkg.in/AlecAivazis/survey.v1"
	"k8s.io/klog"

	"github.com/openshift/odo/pkg/component"
	"github.com/openshift/odo/pkg/odo/cli/ui"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/openshift/odo/pkg/odo/util/validation"
)

// EnterComponentName allows the user to specify the component name in a prompt
func EnterGitOpsRepoURL(context *genericclioptions.Context) string {
	var path string
	prompt := &survey.Input{
		Message: "Enter the GitOps Repo URL",
	}
	err := survey.AskOne(prompt, &path, createComponentNameValidator(context))
	ui.HandleError(err)
	return path
}
func createComponentNameValidator(context *genericclioptions.Context) survey.Validator {
	return func(input interface{}) error {
		if s, ok := input.(string); ok {
			err := validation.ValidateName(s)
			if err != nil {
				return err
			}

			exists, err := component.Exists(context.Client, s, context.Application)
			if err != nil {
				klog.V(4).Info(err)
				return fmt.Errorf("Unable to determine if component '%s' exists or not", s)
			}
			if exists {
				return fmt.Errorf("Component with name '%s' already exists in application '%s'", s, context.Application)
			}

			return nil
		}

		return fmt.Errorf("can only validate strings, got %v", input)
	}
}
