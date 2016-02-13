package controller

import (
	"net/http"

	"github.com/drone/drone/shared/token"
	"github.com/gin-gonic/gin"
	"github.com/mikkeloscar/maze-repo/router/middleware/session"
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
