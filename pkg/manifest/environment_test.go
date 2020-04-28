package manifest

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"

	"github.com/stretchr/testify/assert"
)

func TestEnv(t *testing.T) {

	gitopsPath, cleanUp := fakeGitopsDir(t)
	defer cleanUp()

	envParameters := EnvParameters{
		EnvName: "dev",
		Output:  gitopsPath,
	}
	if err := Env(&envParameters); err != nil {
		t.Fatalf("Env() failed :%s", err)
	}

	wantedPaths := []string{
		"environments/dev/base/kustomization.yaml",
		"environments/dev/base/namespace.yaml",
		"environments/dev/base/rolebinding.yaml",
		"environments/dev/overlays/kustomization.yaml",
	}

	for _, path := range wantedPaths {
		t.Run(fmt.Sprintf("checking path %s already exists", path), func(t *testing.T) {
			assert.FileExists(t, filepath.Join(gitopsPath, path))
		})
	}
}

func fakeGitopsDir(t *testing.T) (string, func()) {
	appFS := afero.NewOsFs()

	tmpDir, cleanUp := makeTempDir(t, appFS)
	gitopsDir := filepath.Join(tmpDir, "gitops")
	err := appFS.Mkdir(gitopsDir, 0755)
	if err != nil {
		t.Fatalf("failed to create gitops directory")
	}
	return gitopsDir, cleanUp
}

func makeTempDir(t *testing.T, f afero.Fs) (string, func()) {
	t.Helper()
	dir, err := ioutil.TempDir(os.TempDir(), "test")
	assertNoError(t, err)
	return dir, func() {
		err := f.RemoveAll(dir)
		assertNoError(t, err)
	}
}

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
