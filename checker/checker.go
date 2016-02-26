package checker

import (
	"fmt"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
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

	updatePkgs, checkPkgs, err := aur.Updates(conf.AUR, r)
	if err != nil {
		return err
	}

	for _, pkgs := range updatePkgs {
		err = c.Remote.EmptyCommit(u,
			r.SourceOwner,
			r.SourceName,
			r.SourceBranch,
			r.BuildBranch,
			fmt.Sprintf("update:%s:aur", strings.Join(pkgs, ",")),
		)
		if err != nil {
			return err
		}
		log.Printf("Making update request for '%s'", strings.Join(pkgs, ", "))
	}

	for _, pkgs := range checkPkgs {
		err = c.Remote.EmptyCommit(u,
			r.SourceOwner,
			r.SourceName,
			r.SourceBranch,
			r.BuildBranch,
			fmt.Sprintf("check:%s:aur", strings.Join(pkgs, ",")),
		)
		if err != nil {
			return err
		}
		log.Printf("Making check request for '%s'", strings.Join(pkgs, ", "))
	}
	return nil
}

// Run runs the checker that checks for package updates in repos.
func (c *Checker) Run() {
	for {
		select {
		case <-time.After(time.Minute * 10):
			// case <-time.After(time.Second * 60):
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
				// last := r.LastCheck.Add(1 * time.Minute)
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
