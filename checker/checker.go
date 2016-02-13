package checker

import (
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/mikkeloscar/maze-repo/model"
	"github.com/mikkeloscar/maze-repo/remote"
	"github.com/mikkeloscar/maze-repo/repo"
	"github.com/mikkeloscar/maze-repo/source/aur"
	"github.com/mikkeloscar/maze-repo/store"
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

	for _, pkg := range pkgs {
		// TODO: configurable branch names
		err = c.Remote.EmptyCommit(u, r.SourceOwner, r.SourceName, "master", "build", fmt.Sprintf("%s:aur", pkg))
		if err != nil {
			return err
		}
	}
	return nil
}

// Run runs the checker that checks for package updates in repos.
func (c *Checker) Run() {
	for {
		select {
		case <-time.After(time.Minute * 5):
			// TODO: maybe run in goroutine
			repos, err := c.Store.Repos().GetRepoList()
			if err != nil {
				log.Errorf("failed to fetch repos from db: %s", err)
				break
			}

			for _, r := range repos {
				user, err := c.Store.Users().Get(r.UserID)
				if err != nil {
					log.Errorf("failed to fetch user from db: %s", err)
					break
				}

				err = c.update(user, &repo.Repo{r})
				if err != nil {
					log.Errorf("failed to request update: %s", err)
					break
				}
			}
		}
	}
}
