package ui

import (
	"gopkg.in/AlecAivazis/survey.v1"

	"github.com/openshift/odo/pkg/odo/cli/ui"
)

// EnterComponentName allows the user to specify the component name in a prompt
func EnterInteractiveCommandLineGitRepo() string {
	var path string
	var prompt *survey.Input
	prompt = &survey.Input{
		Message: "Provide the URL for your GitOps repository",
		Help:    "The GitOps repository stores your GitOps configuration files, including your Openshift Pipelines resources for driving automated deployments and builds.  Please enter a valid git repository e.g. https://github.com/example/myorg.git",
	}

	err := survey.AskOne(prompt, &path, survey.Required)
	ui.HandleError(err)

	return path
}

// OptionBootstrap allows the user to choose if they want to bootstrap or not

func SelectOptionImageRepository() string {
	var path string

	prompt := &survey.Select{
		Message: "Select type of image repository",
		Options: []string{"Openshift Internal repository", "quay.io"},
		Default: "Openshift Internal repository",
	}
	err := survey.AskOne(prompt, &path, survey.Required)
	ui.HandleError(err)
	return path
}

func SelectOptionOverwrite() string {
	var path string

	prompt := &survey.Select{
		Message: "Do you want to overwrite your output path. Select yes or no",
		Options: []string{"yes", "no"},
		Default: "no",
	}
	err := survey.AskOne(prompt, &path, survey.Required)
	ui.HandleError(err)
	return path
}

func SelectOptionCommitStatusTracker() string {
	var path string

	prompt := &survey.Select{
		Message: "Please enter (yes/no) if you desire to use commit-status-tracker",
		Options: []string{"yes", "no"},
		Default: "no",
	}
	err := survey.AskOne(prompt, &path, survey.Required)
	ui.HandleError(err)
	return path
}

func SelectOptionBootstrap() string {
	var path string

	prompt := &survey.Select{
		Message: "Please enter (Bootstrap/init), choose bootstrap if you wish to add a mock service to the gitops repository",
		Options: []string{"Bootstrap", "init"},
		Default: "no",
	}
	err := survey.AskOne(prompt, &path, survey.Required)
	ui.HandleError(err)
	return path
}

func SelectOptionOverWriteCheck() string {
	var path string

	prompt := &survey.Select{
		Message: "Would you like to pass in a different path and try again",
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
// Checks whether the pipelines.yaml is present in the output path specified.
