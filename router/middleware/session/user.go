package session

import (
	"net/http"

	"github.com/drone/drone/shared/token"
	"github.com/gin-gonic/gin"
	"github.com/mikkeloscar/maze-repo/model"
	"github.com/mikkeloscar/maze-repo/store"
)

func User(c *gin.Context) *model.User {
	v, ok := c.Get("user")
	if !ok {
		return nil
	}

	u, ok := v.(*model.User)
	if !ok {
		return nil
	}

	return u
}

func SetUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		var user *model.User

		t, err := token.ParseRequest(c.Request, func(t *token.Token) (string, error) {
			user, err := store.GetUserLogin(c, t.Text)
			return user.Hash, err
		})
		if err == nil {
			c.Set("user", user)

			// if this is a session token (ie not the API token)
			// this means the user is accessing with a web browser,
			// so we should implement CSRF protection measures.
			if t.Kind == token.SessToken {
				err = token.CheckCsrf(c.Request, func(t *token.Token) (string, error) {
					return user.Hash, nil
				})
				// if csrf token validation fails, exit immediately
				// with a not authorized error.
				if err != nil {
					c.AbortWithStatus(http.StatusUnauthorized)
					return
				}
			}
		}
		c.Next()
	}
}

func IsAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		user := User(c)
		if user == nil {
			c.AbortWithStatus(http.StatusUnauthorized)
		} else if !user.Admin {
			c.AbortWithStatus(http.StatusForbidden)
		} else {
			c.Next()
		}
	}
}

func IsUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		user := User(c)
		if user == nil {
			c.AbortWithStatus(http.StatusUnauthorized)
		} else {
			c.Next()
		}
	}
}
