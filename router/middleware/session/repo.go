package session

import (
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"github.com/mikkeloscar/maze/model"
	"github.com/mikkeloscar/maze/repo"
	"github.com/mikkeloscar/maze/store"
)

func Repo(c *gin.Context) *repo.Repo {
	v, ok := c.Get("repo")
	if !ok {
		return nil
	}

	r, ok := v.(*model.Repo)
	if !ok {
		return nil
	}

	return repo.NewRepo(r)
}

func SetRepo() gin.HandlerFunc {
	return func(c *gin.Context) {
		owner := c.Param("owner")
		name := c.Param("name")

		repo, err := store.GetRepoByOwnerName(c, owner, name)
		if err != nil {
			log.Errorf("failed to find repo: %s/%s: %s", owner, name, err)
			c.AbortWithStatus(http.StatusNotFound)
		}

		c.Set("repo", repo)
		c.Next()
	}
}

func RepoPerm(c *gin.Context) *model.Perm {
	v, ok := c.Get("perm")
	if !ok {
		return nil
	}
	u, ok := v.(*model.Perm)
	if !ok {
		return nil
	}
	return u
}

func SetRepoPerm() gin.HandlerFunc {
	return func(c *gin.Context) {
		user := User(c)
		repo := Repo(c)
		perm := &model.Perm{}

		switch {
		case user == nil:
			perm.Pull = false
			perm.Push = false
			perm.Admin = false
		case user.Admin:
			perm.Pull = true
			perm.Push = true
			perm.Admin = true
		case user.ID == repo.UserID:
			perm.Pull = true
			perm.Push = true
			perm.Admin = true
		}

		c.Set("perm", perm)
		c.Next()
	}
}
