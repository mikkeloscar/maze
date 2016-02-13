package session

import (
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"github.com/mikkeloscar/maze-repo/model"
	"github.com/mikkeloscar/maze-repo/repo"
	"github.com/mikkeloscar/maze-repo/store"
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

	return &repo.Repo{r}
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
