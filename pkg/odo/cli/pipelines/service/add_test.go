package service

import (
	"bytes"
	"testing"

	"github.com/openshift/odo/pkg/pipelines"
	"github.com/spf13/cobra"
)

type keyValuePair struct {
	key   string
	value string
}

func TestCompleteAddOptions(t *testing.T) {
	completeTests := []struct {
		name                string
		url                 string
		pipelinesPath       string
		wantURL             string
		wantedPipelinesPath string
	}{
		{"service on GitLab", "https://gitlab.com/test/org", "../test-repo", "https://gitlab.com/test/org.git", "../test-repo/pipelines.yaml"},
		{"service on GitHub", "https://github.com/test/org", "../test-repo/test-repo-2", "https://github.com/test/org.git", "../test-repo/test-repo-2/pipelines.yaml"},
		{"service with no URL", "", "", "", "./pipelines.yaml"},
		{"suffix already present", "https://github.com/test/org.git", "test-repo", "https://github.com/test/org.git", "test-repo/pipelines.yaml"},
	}

	for _, tt := range completeTests {
		t.Run(tt.name, func(rt *testing.T) {
			o := AddServiceOptions{AddServiceOptions: &pipelines.AddServiceOptions{GitRepoURL: tt.url, PipelinesFilePath: tt.pipelinesPath}}
			err := o.Complete("test", &cobra.Command{}, []string{"test", "test/repo"})
			if err != nil {
				rt.Fatal(err)
			}
			if tt.wantURL != o.GitRepoURL && tt.wantedPipelinesPath != o.PipelinesFilePath {
				rt.Fatalf("URL mismatch: got %s, want %s", o.GitRepoURL, tt.wantURL)
			}
		})
	}
}

func TestAddCommandWithMissingParams(t *testing.T) {
	cmdTests := []struct {
		desc    string
		flags   []keyValuePair
		wantErr string
	}{
		{"Missing app-name flag",
			[]keyValuePair{
				flag("service-name", "sample"), flag("git-repo-url", "example/repo"), flag("webhook-secret", "abc123"), flag("env-name", "test")},
			`required flag(s) "app-name" not set`},
		{"Missing service-name flag",
			[]keyValuePair{flag("app-name", "app"),
				flag("git-repo-url", "example/repo"), flag("webhook-secret", "abc123"), flag("env-name", "test")},
			`required flag(s) "service-name" not set`},
		{"Missing env-name flag",
			[]keyValuePair{flag("app-name", "app"),
				flag("service-name", "sample"), flag("git-repo-url", "sample/repo"), flag("webhook-secret", "abc123")},
			`required flag(s) "env-name" not set`},
	}
	for _, tt := range cmdTests {
		t.Run(tt.desc, func(t *testing.T) {
			_, _, err := executeCommand(newCmdAdd("add", "odo pipelines service"), tt.flags...)
			if err.Error() != tt.wantErr {
				t.Errorf("got %s, want %s", err, tt.wantErr)
			}
		})
	}
}

func executeCommand(cmd *cobra.Command, flags ...keyValuePair) (c *cobra.Command, output string, err error) {
	buf := new(bytes.Buffer)
	cmd.SetOutput(buf)
	for _, flag := range flags {
		if err := cmd.Flags().Set(flag.key, flag.value); err != nil {
			return nil, "", err
		}
	}
	c, err = cmd.ExecuteC()
	return c, buf.String(), err
}

func flag(k, v string) keyValuePair {
	return keyValuePair{
		key:   k,
		value: v,
	}
}
