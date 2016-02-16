package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mikkeloscar/maze/store"
)

func GetUser(c *gin.Context) {
	user, err := store.GetUserLogin(c, c.Param("user"))
	if err != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	c.JSON(http.StatusOK, user)
}
