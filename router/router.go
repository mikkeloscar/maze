package router

import (
	"github.com/gin-gonic/gin"
	"github.com/mikkeloscar/maze-repo/controller"
)

// func Load(middleware ...gin.HandlerFunc) http.Handler {
func Load(middleware ...gin.HandlerFunc) *gin.Engine {
	e := gin.Default()
	e.Use(middleware...)

	repos := e.Group("/api/repos/:name")
	{
		repo := repos.Group("")
		{
			repo.POST("/upload/start", controller.PostUploadStart)
			repo.POST("/upload/file/:filename/:sessionid", controller.PostUploadFile)
			repo.POST("/upload/done/:sessionid", controller.PostUploadDone)
		}
	}

	return e
}
