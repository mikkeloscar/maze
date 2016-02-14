package repo

import (
	"fmt"
	"os"
	"path"

	"github.com/drone/drone/shared/envconfig"
)

// RepoStorage defines the repo storage basepath.
var RepoStorage = ""

// LoadRepoStorage checks if the path set by REPO_STORAGE is available and
// tries to create it if it doesn't exist.
func LoadRepoStorage(env envconfig.Env) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	RepoStorage = env.String("REPO_STORAGE", path.Join(wd, "_repo_storage"))

	f, err := os.Stat(RepoStorage)
	if err != nil {
		if os.IsNotExist(err) {
			return os.MkdirAll(RepoStorage, 0755)
		}
		return err
	}

	if !f.IsDir() {
		return fmt.Errorf("repo storage path %s is not a directory", RepoStorage)
	}

	return nil
}
