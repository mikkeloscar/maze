package checker

import (
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/mikkeloscar/maze/common/util"
	"github.com/mikkeloscar/maze/model"
	"github.com/mikkeloscar/maze/remote"
	"github.com/mikkeloscar/maze/repo"
	"github.com/mikkeloscar/maze/source/aur"
	"github.com/mikkeloscar/maze/store"
)

// Checker can check for package updates in repos.
type Checker struct {
	Remote remote.Remote
	Store  store.Store
}

// trigger update builds for new packages.
func (c *Checker) update(u *model.User, r *repo.Repo) error {
	conf, err := c.Remote.GetConfig(u, r.SourceOwner, r.SourceName, "packages.yml")
	if err != nil {
		return err
	}

	pkgs, err := aur.Updates(conf.AUR, r)
	if err != nil {
		return err
	}

	checkPkgs := make([]string, 0)
	for _, pkg := range conf.AUR {
		if !util.StrContains(pkg, pkgs) && util.IsDevel(pkg) {
			checkPkgs = append(checkPkgs, pkg)
		}
	}

	for _, pkg := range pkgs {
		// TODO: configurable branch names
		err = c.Remote.EmptyCommit(u, r.SourceOwner, r.SourceName, "master", "build", fmt.Sprintf("update:%s:aur", pkg))
		if err != nil {
			return err
		}
		log.Printf("Making update request for '%s'", pkg)
	}

	for _, pkg := range checkPkgs {
		// TODO: configurable branch names
		err = c.Remote.EmptyCommit(u, r.SourceOwner, r.SourceName, "master", "build", fmt.Sprintf("check:%s:aur", pkg))
		if err != nil {
			return err
		}
		log.Printf("Making check request for '%s'", pkg)
	}
	return nil
}

// Run runs the checker that checks for package updates in repos.
func (c *Checker) Run() {
	for {
		select {
		case <-time.After(time.Minute * 10):
			// case <-time.After(time.Second * 10):
			// TODO: maybe run in goroutine
			repos, err := c.Store.Repos().GetRepoList()
			if err != nil {
				log.Errorf("failed to fetch repos from db: %s", err)
				break
			}

			if len(repos) > 0 {
				log.Info("Checking for package updates for all repos")
			}

			for _, r := range repos {
				// only check for updates if last check was
				// more than an hour ago
				last := r.LastCheck.Add(1 * time.Hour)
				if time.Now().UTC().Before(last) {
					continue
				}

				user, err := c.Store.Users().Get(r.UserID)
				if err != nil {
					log.Errorf("failed to fetch user from db: %s", err)
					break
				}

				err = c.update(user, repo.NewRepo(r, repo.RepoStorage))
				if err != nil {
					log.Errorf("failed to request update: %s", err)
					break
				}

				// update lastCheck
				r.LastCheck = time.Now().UTC()
				c.Store.Repos().Update(r)
			}
		}
	}
}
