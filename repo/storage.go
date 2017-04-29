package repo

import (
	"fmt"
	"os"
)

// RepoStorage defines the repo storage basepath.
var RepoStorage = ""

// LoadRepoStorage checks if the path set by REPO_STORAGE is available and
// tries to create it if it doesn't exist.
func LoadRepoStorage() error {
	RepoStorage = os.Getenv("REPO_STORAGE")

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
