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

// OptionBootstrap allows the user to choose if they want to bootstrap or not

func SelectOption(message string) string {
	var path string

	prompt := &survey.Select{
		Message: message,
		Options: []string{"yes", "no"},
		Default: "no",
	}
	err := survey.AskOne(prompt, &path, survey.Required)
	ui.HandleError(err)
	return path
}

// // Proceed displays a given message and asks the user if they want to proceed using the optionally specified Stdio instance (useful
// // for testing purposes)
// func Proceed(message string, stdio ...terminal.Stdio) bool {
// 	var response bool
// 	prompt := &survey.Confirm{
// 		Message: message,
// 	}

// 	if len(stdio) == 1 {
// 		prompt.WithStdio(stdio[0])
// 	}

// 	err := survey.AskOne(prompt, &response, survey.Required)
// 	HandleError(err)

// 	return response
// }
