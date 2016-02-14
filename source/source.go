package source

import "github.com/mikkeloscar/maze/repo"

type Source interface {
	// Updates takes a list of packagenames and returns a list of packages
	// where a new version is available based on the provided respository.
	Updates([]string, *repo.Repo) ([]string, error)
}

func Updates(typ string, pkgs []string, repo *repo.Repo) ([]string, error) {
	switch typ {
	case "aur":
		return aur.Updates(pkgs, repo)
	}

	return nil, fmt.Errorf("invalid source type")
}
