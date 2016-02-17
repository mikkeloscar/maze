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
func Updates(pkgs []string, repo *repo.Repo) ([]*source.SourcePkg, []*source.SourcePkg, error) {
	deps := make(map[string]*source.SourcePkg)
	err := getDeps(pkgs, deps)
	if err != nil {
		return nil, nil, err
	}

	updates := make([]*source.SourcePkg, 0)
	checks := make([]*source.SourcePkg, 0)

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
func getDeps(pkgs []string, updates map[string]*source.SourcePkg) error {
	pkgsInfo, err := aur.Multiinfo(pkgs)
	if err != nil {
		return err
	}

	for _, pkg := range pkgsInfo {
		updates[pkg.Name] = &source.SourcePkg{
			Name:    pkg.Name,
			Version: pkg.Version,
		}

		// TODO: maybe add optdepends
		depends := make([]string, 0, len(pkg.Depends)+len(pkg.MakeDepends))
		depends = append(depends, pkg.Depends...)
		depends = append(depends, pkg.MakeDepends...)
		getDeps(depends, updates)
	}

	return nil
}
