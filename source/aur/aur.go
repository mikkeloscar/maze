package aur

import (
	"github.com/mikkeloscar/aur"
	"github.com/mikkeloscar/gopkgbuild"
	"github.com/mikkeloscar/maze/common/util"
	"github.com/mikkeloscar/maze/repo"
)

// Updates check for updated packages based on a list of packages and a
// repository. Returns a list of packages with updates.
func Updates(pkgs []string, repo *repo.Repo) ([][]string, [][]string, error) {
	deps := make(map[string]*depNode)
	err := getDeps(pkgs, nil, deps)
	if err != nil {
		return nil, nil, err
	}

	groups := groupDeps(deps)

	var updates [][]string
	var checks [][]string

	for _, group := range groups {
		var updatesGroup []string
		var checksGroup []string
		for name, version := range group {
			compVersion, err := pkgbuild.NewCompleteVersion(version)
			if err != nil {
				return nil, nil, err
			}

			new, err := repo.IsNew(name, "any", *compVersion)
			if err != nil {
				return nil, nil, err
			}

			if new {
				updatesGroup = append(updatesGroup, name)
			}

			if util.IsDevel(name) && !new {
				checksGroup = append(checksGroup, name)
			}
		}

		if len(updatesGroup) > 0 {
			updates = append(updates, updatesGroup)
		}

		if len(checksGroup) > 0 {
			checks = append(checks, checksGroup)
		}
	}

	return updates, checks, nil
}

type depNode struct {
	name     string
	version  string
	parents  map[string]*depNode
	children map[string]*depNode
}

// query the AUR for build deps to packages.
func getDeps(pkgs []string, curr *depNode, updates map[string]*depNode) error {
	var err error
	pkgsInfo, err := aur.Info(pkgs)
	if err != nil {
		return err
	}

	for _, pkg := range pkgsInfo {
		var pkgNode *depNode

		pkgNode, ok := updates[pkg.Name]
		if !ok {
			pkgNode = &depNode{
				name:     pkg.Name,
				version:  pkg.Version,
				parents:  make(map[string]*depNode),
				children: make(map[string]*depNode),
			}
			updates[pkg.Name] = pkgNode
		}

		if curr != nil {
			curr.parents[pkgNode.name] = pkgNode
			pkgNode.children[curr.name] = curr
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
		getDeps(depends, pkgNode, updates)
	}

	return nil
}

func groupDeps(pkgs map[string]*depNode) []map[string]string {
	var groups []map[string]string

	for _, d := range pkgs {
		if len(d.children) != 0 {
			continue
		}

		group := make(map[string]string)
		followNode(d, group, pkgs)
		if len(group) > 0 {
			groups = append(groups, group)
		}
	}

	return groups
}

func followNode(n *depNode, g map[string]string, table map[string]*depNode) {
	if len(n.children) == 0 {
		g[n.name] = n.version
		delete(table, n.name)
		for _, p := range n.parents {
			g[p.name] = p.version
			if len(p.children) > 1 {
				delete(p.children, n.name)
				followNode(p, g, table)
			}
		}
		return
	}

	for _, c := range n.children {
		followNode(c, g, table)
	}
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
