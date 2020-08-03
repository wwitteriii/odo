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

func EnterInteractiveCommandLineStatusTrackerAccessToken() string{
	var path string
	prompt = &survey.Password{
		Message: "Please provide a token used to authenticate API calls to push commit-status updates to your Git hosting service",
		Help: "commit-status-tracker reports the completion status of OpenShift pipeline runs to your Git hosting status on success or failure, this token will be encrypted as a secret in your cluster.\n If you are using Github, please see here for how to generate a token https://docs.github.com/en/github/authenticating-to-github/creating-a-personal-access-token\nIf you are using GitLab, please see here for how to generate a token https://docs.gitlab.com/ee/user/profile/personal_access_tokens.html",
	}
	err := survey.AskOne(prompt, &path, survey.Required)
	ui.HandleError(err)
	return path
}

//
func EnterInteractiveCommandLinePrefix() string{
	var path string
	prompt := &survey.Input{
		Message: "Add a prefix to the environment names(dev, stage,cicd etc.) to distinguish and identify individual environments?",
		Help: "The prefix helps differentiate between the different namespaces on the cluster, the default namespace cicd will appear as test-cicd if the prefix passed is test.",
	}
	err := survey.AskOne(prompt, &path, nil)
	ui.HandleError(err)
	return path
}
//
func EnterInteractiveCommandLineServiceRepoURL() string {
	var path string
	prompt := &survey.Input{
		Message: "Provide the URL for your Service repository e.g. https://github.com/organisation/service.git",
		Help: "The repository name where the source code of your service is situated",
	}
	err := survey.AskOne(prompt, &path, survey.Required)
	ui.HandleError(err)
	return path
}

// EnterInteractiveCommandLineServiceWebhookSecret function for Service webhook secret
func EnterInteractiveCommandLineServiceWebhookSecret() string {
	var path string
	prompt := &survey.Input{
		Message: " Provide a secret that we can use to authenticate incoming hooks from your Git hosting service for the Service repository. (if not provided, it will be auto-generated)",
		Default: "",
		Help: "The webhook secret is a secure string you plan to use to authenticate pull/push requests to the version control system of your choice, this secure string will be added to the webhook sealed secret created to enhance security. Choose a secure string of your choice for this field."
	}
	err := survey.AskOne(prompt, &path, nil)
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
