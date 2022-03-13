package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mikkeloscar/maze/pkg/token"
	"github.com/mikkeloscar/maze/router/middleware/session"
)

func GetSelf(c *gin.Context) {
	c.JSON(http.StatusOK, session.User(c))
}

func PostToken(c *gin.Context) {
	user := session.User(c)

	token := token.New(token.UserToken, user.Login)
	tokenstr, err := token.Sign(user.Hash)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.String(http.StatusOK, tokenstr)
}
