package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mikkeloscar/maze-repo/store"
)

func GetUser(c *gin.Context) {
	user, err := store.GetUserLogin(c, c.Param("login"))
	if err != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	c.JSON(http.StatusOK, user)
}
