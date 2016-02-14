package controller

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/drone/drone/shared/crypto"
	"github.com/gin-gonic/gin"
	"github.com/mikkeloscar/maze/remote"
	"github.com/mikkeloscar/maze/router/middleware/session"
	"github.com/mikkeloscar/maze/store"
	"github.com/satori/go.uuid"
)

var sessions = map[string][]string{}

func getSessionID() string {
	for {
		u := uuid.NewV4().String()
		if _, ok := sessions[u]; !ok {
			sessions[u] = make([]string, 0)
			return u
		}
	}
}

type MetaPkg struct {
	Package   string `json:"package" binding:"required"`
	Signature string `json:"signature" binding:"required"`
}

func PostUploadStart(c *gin.Context) {
	// name := c.Param("name")
	// repository := repo.GetByName(name)
	// repo := session.Repo(c)

	c.JSON(http.StatusOK, gin.H{
		"session_id": getSessionID(),
	})
}

func PostUploadFile(c *gin.Context) {
	pkg := c.Param("filename")
	sessionID := c.Param("sessionid")
	repo := session.Repo(c)

	if _, ok := sessions[sessionID]; !ok {
		c.AbortWithStatus(http.StatusNotFound)
	}

	// TODO check valid filename
	// TODO check valid file content

	new, err := repo.IsNewFilename(pkg)
	if err != nil {
		c.AbortWithError(400, err)
		return
	}

	if !new {
		c.AbortWithStatus(208)
		return
	}

	pkg = path.Join(repo.Path, pkg)

	f, err := os.Create(pkg)
	if err != nil {
		log.Errorf("failed to create file %s: %s", pkg, err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	_, err = io.Copy(f, c.Request.Body)
	if err != nil {
		log.Errorf("failed to write data: %s", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	sessions[sessionID] = append(sessions[sessionID], pkg)

	c.Writer.WriteHeader(http.StatusOK)
}

func PostUploadDone(c *gin.Context) {
	sessionID := c.Param("sessionid")
	repo := session.Repo(c)

	if v, ok := sessions[sessionID]; ok {
		delete(sessions, sessionID)
		err := repo.Add(v)
		if err != nil {
			log.Errorf("failed to add package '%s' to repository '%s': %s", v, repo.Name, err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		c.Writer.WriteHeader(http.StatusOK)
		return
	}

	c.AbortWithStatus(http.StatusNotFound)
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
	r.LastCheck = time.Now().Add(-1 * time.Hour)
	r.Hash = crypto.Rand()

	err = store.CreateRepo(c, r)
	if err != nil {
		log.Errorf("failed to add repo: %s", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, r)
}

func GetRepo(c *gin.Context) {
	c.JSON(http.StatusOK, session.Repo(c))
}
