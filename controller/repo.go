package controller

import (
	"io"
	"os"
	"path"

	"github.com/gin-gonic/gin"
	"github.com/mikkeloscar/maze-repo/repo"
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
	name := c.Param("name")
	repository := repo.GetByName(name)

	if repository == nil {
		c.AbortWithStatus(404)
		return
	}

	c.JSON(200, gin.H{
		"session_id": getSessionID(),
	})
}

func PostUploadFile(c *gin.Context) {
	name := c.Param("name")
	pkg := c.Param("filename")
	sessionID := c.Param("sessionid")
	repository := repo.GetByName(name)

	if repository == nil {
		c.AbortWithStatus(404)
		return
	}

	if _, ok := sessions[sessionID]; !ok {
		c.AbortWithStatus(404)
	}

	// TODO check valid filename

	new, err := repository.IsNewFilename(pkg)
	if err != nil {
		c.AbortWithError(400, err)
		return
	}

	if !new {
		c.AbortWithStatus(208)
		return
	}

	pkg = path.Join(repository.Path, pkg)

	f, err := os.Create(pkg)
	if err != nil {
		// TODO: log error.
		c.AbortWithStatus(500)
		return
	}

	_, err = io.Copy(f, c.Request.Body)
	if err != nil {
		// TODO: log error.
		c.AbortWithStatus(500)
		return
	}

	sessions[sessionID] = append(sessions[sessionID], pkg)

	c.Writer.WriteHeader(200)
}

func PostUploadDone(c *gin.Context) {
	name := c.Param("name")
	sessionID := c.Param("sessionid")
	repository := repo.GetByName(name)

	if repository == nil {
		c.AbortWithStatus(404)
		return
	}

	if v, ok := sessions[sessionID]; ok {
		delete(sessions, sessionID)
		err := repository.Add(v)
		if err != nil {
			// TODO: log error.
			c.AbortWithStatus(500)
			return
		}
		c.Writer.WriteHeader(200)
		return
	}

	c.AbortWithStatus(404)
}
