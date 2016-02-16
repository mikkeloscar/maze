package controller

import (
	"fmt"
	"net/http"
	"path"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/drone/drone/shared/crypto"
	"github.com/gin-gonic/gin"
	"github.com/mikkeloscar/maze/remote"
	"github.com/mikkeloscar/maze/repo"
	"github.com/mikkeloscar/maze/router/middleware/session"
	"github.com/mikkeloscar/maze/store"
)

func ServeRepoFile(c *gin.Context) {
	repo := session.Repo(c)
	file := c.Param("file")
	c.File(path.Join(repo.Path(), file))
}

func splitRepoName(source string) (string, string, error) {
	split := strings.Split(source, "/")
	if len(split) != 2 {
		return "", "", fmt.Errorf("invalid repo format: %s", source)
	}

	return split[0], split[1], nil
}

func PostRepo(c *gin.Context) {
	remote := remote.FromContext(c)
	user := session.User(c)
	owner := c.Param("owner")
	name := c.Param("name")

	if !repo.ValidRepoName(name) {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if owner != user.Login {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	in := struct {
		SourceRepo string `json:"source_repo" binding:"required"`
	}{}
	err := c.BindJSON(&in)
	if err != nil {
		log.Errorf("failed to parse request body: %s", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	sourceOwner, sourceName, err := splitRepoName(in.SourceRepo)
	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	// error if the repository already exists
	_r, err := store.GetRepoByOwnerName(c, owner, name)
	if _r != nil {
		log.Errorf("unable to add repo: %s/%s, already exists", owner, name)
		c.AbortWithStatus(http.StatusConflict)
		return
	}

	// Fetch source repo
	r, err := remote.Repo(user, sourceOwner, sourceName)
	if err != nil {
		log.Errorf("unable to get repo: %s/%s: %s", sourceOwner, sourceName, err)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	p, err := remote.Perm(user, sourceOwner, sourceName)
	if err != nil {
		log.Errorf("unable to get repo permission for: %s/%s: %s", sourceOwner, sourceName, err)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	if !(p.Admin || (p.Pull && p.Push)) {
		log.Errorf("pull/push access required")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	// TODO: make branches configurable
	err = remote.SetupBranch(user, sourceOwner, sourceName, "master", "build")
	if err != nil {
		log.Errorf("failed to setup build branch: %s", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	r.UserID = user.ID
	r.Owner = owner
	r.Name = name
	r.LastCheck = time.Now().UTC().Add(-1 * time.Hour)
	r.Hash = crypto.Rand()

	fsRepo := repo.NewRepo(r, repo.RepoStorage)

	err = fsRepo.InitDir()
	if err != nil {
		log.Errorf("failed to create repo storage path '%s' on disk: %s", fsRepo.Path(), err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	err = store.CreateRepo(c, r)
	if err != nil {
		log.Errorf("failed to add repo: %s", err)
		err = fsRepo.ClearPath()
		if err != nil {
			log.Errorf("failed to cleanup repo path '%s': %s", fsRepo.Path(), err)
		}
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, r)
}

func GetRepo(c *gin.Context) {
	c.JSON(http.StatusOK, session.Repo(c))
}

func PatchRepo(c *gin.Context) {
	r := session.Repo(c)

	in := struct {
		SourceOwner *string `json:"source_owner,omitempty"`
		SourceName  *string `json:"source_name,omitempty"`
		Name        *string `json:"name,omitempty"`
	}{}

	err := c.BindJSON(&in)
	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if in.SourceOwner != nil {
		r.SourceOwner = *in.SourceOwner
	}

	if in.SourceName != nil {
		r.SourceName = *in.SourceName
	}

	if in.Name != nil {
		r.Name = *in.Name
		if !repo.ValidRepoName(r.Name) {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
	}

	err = store.UpdateRepo(c, r.Repo)
	if err != nil {
		log.Errorf("failed to update repo '%s/%s': %s", r.Owner, r.Name, err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, r)
}

func DeleteRepo(c *gin.Context) {
	repo := session.Repo(c)

	err := repo.ClearPath()
	if err != nil {
		log.Errorf("failed to delete repo storage '%s/%s': %s", repo.Owner, repo.Name, err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	err = store.DeleteRepo(c, repo.Repo)
	if err != nil {
		log.Errorf("failed to remove repo db entry '%s/%s': %s", repo.Owner, repo.Name, err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusOK)
}

func GetRepoPackages(c *gin.Context) {
	pkgs, err := session.Repo(c).Packages(false)
	if err != nil {
		log.Errorf("Failed to get repo packages: %s", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, pkgs)
}

func GetRepoPackage(c *gin.Context) {
	pkgname := c.Param("package")
	pkg, err := session.Repo(c).Package(pkgname, false)
	if err != nil {
		log.Errorf("Failed to get repo package '%s': %s", pkgname, err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	if pkg == nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	c.JSON(http.StatusOK, pkg)
}

func GetRepoPackageFiles(c *gin.Context) {
	pkgname := c.Param("package")
	pkg, err := session.Repo(c).Package(pkgname, true)
	if err != nil {
		log.Errorf("Failed to get repo package '%s': %s", pkgname, err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	if pkg == nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	c.JSON(http.StatusOK, pkg.Files)
}
