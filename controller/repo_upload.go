package controller

import (
	"io"
	"net/http"
	"os"
	"path"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"github.com/mikkeloscar/maze/checker"
	"github.com/mikkeloscar/maze/router/middleware/session"
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
	// TODO: clear session on error

	new, err := repo.IsNewFilename(pkg)
	if err != nil {
		c.AbortWithError(400, err)
		return
	}

	if !new {
		c.AbortWithStatus(208)
		return
	}

	pkg = path.Join(repo.Path(), pkg)

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
	state := checker.FromContext(c)

	if pkgs, ok := sessions[sessionID]; ok {
		delete(sessions, sessionID)
		err := repo.Add(pkgs)
		if err != nil {
			log.Errorf("failed to add packages '%s' to repository '%s': %s", strings.Join(pkgs, ", "), repo.Name, err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		// clear packages from checker state
		// TODO: check pkg name (maybe it's the full path here?)
		for _, pkg := range pkgs {
			state.ClearPkg(pkg, repo.Owner, repo.Name)
		}
		c.Writer.WriteHeader(http.StatusOK)
		return
	}

	c.AbortWithStatus(http.StatusNotFound)
}
