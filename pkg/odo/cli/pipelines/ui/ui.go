package ui

import (
	"gopkg.in/AlecAivazis/survey.v1"

	"github.com/openshift/odo/pkg/odo/cli/ui"
)

// EnterComponentName allows the user to specify the component name in a prompt
func EnterInteractiveCommandLine(message, defaultValue string, required bool) string {
	var path string
	var prompt *survey.Input
	if defaultValue == "" {
		prompt = &survey.Input{
			Message: message,
		}
	} else {
		prompt = &survey.Input{
			Message: message,
			Default: defaultValue,
		}
	}
	if required {
		err := survey.AskOne(prompt, &path, survey.Required)
		ui.HandleError(err)

	} else {
		err := survey.AskOne(prompt, &path, nil)
		ui.HandleError(err)
	}
	return path
}
