package aur

import (
	"github.com/mikkeloscar/aur"
	"github.com/mikkeloscar/gopkgbuild"
	"github.com/mikkeloscar/maze/common/util"
	"github.com/mikkeloscar/maze/repo"
)

// Updates check for updated packages based on a list of packages and a
// repository. Returns a list of packages with updates.
func Updates(pkgs []string, repo *repo.Repo) ([]string, []string, error) {
	deps := make(map[string]string)
	err := getDeps(pkgs, deps)
	if err != nil {
		return nil, nil, err
	}

	updates := make([]string, 0)
	checks := make([]string, 0)

	for name, version := range deps {
		compVersion, err := pkgbuild.NewCompleteVersion(version)
		if err != nil {
			return nil, nil, err
		}

		new, err := repo.IsNew(name, *compVersion)
		if err != nil {
			return nil, nil, err
		}

		if new {
			updates = append(updates, name)
		}

		if util.IsDevel(name) && !new {
			checks = append(checks, name)
		}
	}

	return updates, checks, nil
}

// query the AUR for build deps to packages.
func getDeps(pkgs []string, updates map[string]string) error {
	pkgsInfo, err := aur.Multiinfo(pkgs)
	if err != nil {
		return err
	}

	for _, pkg := range pkgsInfo {
		updates[pkg.Name] = pkg.Version

		// TODO: maybe add optdepends
		depends := make([]string, 0, len(pkg.Depends)+len(pkg.MakeDepends))
		depends = append(depends, pkg.Depends...)
		depends = append(depends, pkg.MakeDepends...)
		getDeps(depends, updates)
	}

	return nil
}
