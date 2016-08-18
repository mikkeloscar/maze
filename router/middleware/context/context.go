package context

import (
	"github.com/gin-gonic/gin"
	"github.com/mikkeloscar/maze/checker"
	"github.com/mikkeloscar/maze/remote"
	"github.com/mikkeloscar/maze/store"
)

func SetStore(s store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		store.ToContext(c, s)
		c.Next()
	}
}

func SetRemote(remote remote.Remote) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("remote", remote)
		c.Next()
	}
}

func Remote(c *gin.Context) remote.Remote {
	return c.MustGet("remote").(remote.Remote)
}

func SetState(state *checker.State) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("state", state)
		c.Next()
	}
}
