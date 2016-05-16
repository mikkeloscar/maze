package controller

import (
	"encoding/base32"
	"fmt"
	"net/http"
	"path"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/securecookie"
	"github.com/mikkeloscar/maze/common/util"
	"github.com/mikkeloscar/maze/remote"
	"github.com/mikkeloscar/maze/repo"
	"github.com/mikkeloscar/maze/router/middleware/session"
	"github.com/mikkeloscar/maze/store"
)

func ServeRepoFile(c *gin.Context) {
	repo := session.Repo(c)
	arch := c.Param("arch")
	file := c.Param("file")
	c.File(path.Join(repo.PathDeep(arch), file))
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
		SourceRepo   *string   `json:"source_repo" binding:"required"`
		SourceBranch *string   `json:"source_branch,omitempty"`
		BuildBranch  *string   `json:"build_branch,omitempty"`
		Archs        *[]string `json:"archs,omitempty"`
		Private      *bool     `json:"private,omitempty"`
	}{}
	err := c.BindJSON(&in)
	if err != nil {
		log.Errorf("failed to parse request body: %s", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if in.Private == nil {
		private := false
		in.Private = &private
	}

	if in.Archs == nil || len(*in.Archs) == 0 {
		defArch := []string{"x86_64"}
		in.Archs = &defArch
	}

	if !repo.ValidArchs(*in.Archs) {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	sourceOwner, sourceName, err := splitRepoName(*in.SourceRepo)
	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if in.SourceBranch == nil {
		sourceBranch := "master"
		in.SourceBranch = &sourceBranch
	}

	if in.BuildBranch == nil {
		buildBranch := "build"
		in.BuildBranch = &buildBranch
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

	if !(p.Admin || (p.Read && p.Write)) {
		log.Errorf("pull/push access required")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	err = remote.SetupBranch(user, sourceOwner, sourceName, *in.SourceBranch, *in.BuildBranch)
	if err != nil {
		log.Errorf("failed to setup build branch: %s", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	r.UserID = user.ID
	r.Owner = owner
	r.Name = name
	r.Private = *in.Private
	r.SourceBranch = *in.SourceBranch
	r.BuildBranch = *in.BuildBranch
	r.LastCheck = time.Now().UTC().Add(-1 * time.Hour)
	r.Hash = base32.StdEncoding.EncodeToString(
		securecookie.GenerateRandomKey(32),
	)

	fsRepo := repo.NewRepo(r, repo.RepoStorage)

	err = fsRepo.InitDir()
	if err != nil {
		log.Errorf("failed to create repo storage path '%s' on disk: %s", fsRepo.Path(), err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	err = fsRepo.InitEmptyDBs()
	if err != nil {
		log.Errorf("failed to initialize empty dbs: %s", err)
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
		SourceOwner  *string `json:"source_owner,omitempty"`
		SourceName   *string `json:"source_name,omitempty"`
		SourceBranch *string `json:"source_branch,omitempty"`
		BuildBranch  *string `json:"build_branch,omitempty"`
		Name         *string `json:"name,omitempty"`
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

	if in.SourceBranch != nil {
		r.SourceBranch = *in.SourceBranch
	}

	if in.BuildBranch != nil {
		r.BuildBranch = *in.BuildBranch
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
	repo := session.Repo(c)
	arch := c.Param("arch")

	if !util.StrContains(arch, repo.Archs) {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	pkgs, err := repo.Packages(arch, false)
	if err != nil {
		log.Errorf("Failed to get repo packages: %s", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, pkgs)
}

func GetRepoPackage(c *gin.Context) {
	repo := session.Repo(c)
	pkgname := c.Param("package")
	arch := c.Param("arch")

	if !util.StrContains(arch, repo.Archs) {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	pkg, err := repo.Package(pkgname, arch, false)
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
	repo := session.Repo(c)
	pkgname := c.Param("package")
	arch := c.Param("arch")

	if !util.StrContains(arch, repo.Archs) {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	pkg, err := repo.Package(pkgname, arch, true)
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

func DeleteRepoPackage(c *gin.Context) {
	repo := session.Repo(c)
	pkgname := c.Param("package")
	arch := c.Param("arch")

	if !util.StrContains(arch, repo.Archs) {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	pkg, err := repo.Package(pkgname, arch, true)
	if err != nil {
		log.Errorf("Failed to get repo package '%s': %s", pkgname, err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	if pkg == nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	err = repo.Remove([]string{pkgname}, arch)
	if err != nil {
		log.Errorf("Failed to remove repo package '%s': %s", pkgname, err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusOK)
}
