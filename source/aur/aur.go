package aur

import (
	"github.com/mikkeloscar/aur"
	"github.com/mikkeloscar/gopkgbuild"
	"github.com/mikkeloscar/maze/common/util"
	"github.com/mikkeloscar/maze/repo"
	"github.com/mikkeloscar/maze/source"
)

// Updates check for updated packages based on a list of packages and a
// repository. Returns a list of packages with updates.
func Updates(pkgs []string, repo *repo.Repo) ([]*source.Pkg, []*source.Pkg, error) {
	deps := make(map[string]*source.Pkg)
	err := getDeps(pkgs, deps)
	if err != nil {
		return nil, nil, err
	}

	var updates []*source.Pkg
	var checks []*source.Pkg

	for name, pkg := range deps {
		compVersion, err := pkgbuild.NewCompleteVersion(pkg.Version)
		if err != nil {
			return nil, nil, err
		}

		new, err := repo.IsNew(name, "any", *compVersion)
		if err != nil {
			return nil, nil, err
		}

		if new {
			updates = append(updates, pkg)
		}

		if util.IsDevel(name) && !new {
			checks = append(checks, pkg)
		}
	}

	return updates, checks, nil
}

// query the AUR for build deps to packages.
func getDeps(pkgs []string, updates map[string]*source.Pkg) error {
	pkgsInfo, err := aur.Info(pkgs)
	if err != nil {
		return err
	}

	for _, pkg := range pkgsInfo {
		updates[pkg.Name] = &source.Pkg{
			Name:    pkg.Name,
			Version: pkg.Version,
		}

		// TODO: maybe add optdepends
		depends := make([]string, 0, len(pkg.Depends)+len(pkg.MakeDepends))
		err := addDeps(&depends, pkg.Depends)
		if err != nil {
			return err
		}
		err = addDeps(&depends, pkg.MakeDepends)
		if err != nil {
			return err
		}
		getDeps(depends, updates)
	}

	return nil
}

// parses a string slice of dependencies and adds them to the combinedDepends
// slice.
func addDeps(combinedDepends *[]string, deps []string) error {
	parsedDeps, err := pkgbuild.ParseDeps(deps)
	if err != nil {
		return err
	}

	for _, dep := range parsedDeps {
		*combinedDepends = append(*combinedDepends, dep.Name)
	}

	return nil
}
