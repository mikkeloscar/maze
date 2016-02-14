package router

import (
	"github.com/gin-gonic/gin"
	"github.com/mikkeloscar/maze/controller"
	"github.com/mikkeloscar/maze/router/middleware/session"
)

// func Load(middleware ...gin.HandlerFunc) http.Handler {
func Load(middleware ...gin.HandlerFunc) *gin.Engine {
	e := gin.Default()
	e.Use(middleware...)
	e.Use(session.SetUser())

	e.GET("/logout", controller.GetLogout)

	repos := e.Group("/api/repos/:owner/:name")
	{
		repos.POST("", session.IsUser(), controller.PostRepo)

		repo := repos.Group("")
		{
			repo.Use(session.SetRepo())
			repo.Use(session.SetRepoPerm())
			// TODO: more advanced permissions

			repo.GET("", controller.GetRepo)

			upload := repo.Group("/upload")
			{
				upload.POST("/start", controller.PostUploadStart)
				upload.POST("/file/:filename/:sessionid", controller.PostUploadFile)
				upload.POST("/done/:sessionid", controller.PostUploadDone)
			}
		}
	}

	user := e.Group("/api/user")
	{
		user.Use(session.IsUser())
		user.GET("", controller.GetSelf)
		user.POST("/token", controller.PostToken)
	}

	users := e.Group("/api/users")
	{
		user.Use(session.IsAdmin())
		users.GET("/:login", controller.GetUser)
	}

	auth := e.Group("/authorize")
	{
		auth.GET("", controller.GetLogin)
		auth.POST("", controller.GetLogin)
	}

	return e
}
