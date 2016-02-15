package aur

import (
	"strings"

	"github.com/mikkeloscar/aur"
	"github.com/mikkeloscar/gopkgbuild"
	"github.com/mikkeloscar/maze/repo"
)

// Updates check for updated packages based on a list of packages and a
// repository. Returns a list of packages with updates.
func Updates(pkgs []string, repo *repo.Repo) ([]string, error) {
	deps := make(map[string]string)
	err := getDeps(pkgs, deps)
	if err != nil {
		return nil, err
	}

	updates := make([]string, 0)

	for name, version := range deps {
		compVersion, err := pkgbuild.NewCompleteVersion(version)
		if err != nil {
			return nil, err
		}

		new, err := repo.IsNew(name, *compVersion)
		if err != nil {
			return nil, err
		}

		if new || isDevel(name) {
			updates = append(updates, name)
		}
	}

	return updates, nil
}

// returns true if the pkg is a devel package (ends on -{bzr,git,hg,svn}).
func isDevel(pkg string) bool {
	if strings.HasSuffix(pkg, "-git") {
		return true
	}

	if strings.HasSuffix(pkg, "-svn") {
		return true
	}

	if strings.HasSuffix(pkg, "-hg") {
		return true
	}

	if strings.HasSuffix(pkg, "-bzr") {
		return true
	}

	return false
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
