package router

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mikkeloscar/maze/controller"
	"github.com/mikkeloscar/maze/router/middleware/session"
)

func Load(middleware ...gin.HandlerFunc) http.Handler {
	e := gin.Default()
	e.Use(middleware...)
	e.Use(session.SetUser())

	e.GET("/logout", controller.GetLogout)

	repo := e.Group("/repos/:owner/:name")
	{
		repo.Use(session.SetRepo())
		repo.GET("/:file", controller.ServeRepoFile)
	}

	repos := e.Group("/api/repos/:owner/:name")
	{
		repos.POST("", session.IsUser(), controller.PostRepo)

		repo := repos.Group("")
		{
			repo.Use(session.SetRepo())
			repo.Use(session.SetRepoPerm())
			// TODO: more advanced permissions

			repo.GET("", controller.GetRepo)
			// TODO: add permissions
			repo.PATCH("", controller.PatchRepo)
			repo.DELETE("", controller.DeleteRepo)

			packages := repo.Group("/packages")
			{
				packages.GET("", controller.GetRepoPackages)
				packages.GET("/:package", controller.GetRepoPackage)
				packages.GET("/:package/files", controller.GetRepoPackageFiles)
			}

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
		// TODO: not secure!!! temp hack while we don't have an UI.
		user.GET("/token", controller.PostToken)
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

	return normalize(e)
}

// normalize is a helper function to work around the following
// issue with gin. https://github.com/gin-gonic/gin/issues/388
func normalize(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		parts := strings.Split(r.URL.Path, "/")[1:]
		switch parts[0] {
		case "repos", "api", "login", "logout", "", "authorize":
			// no-op
		default:

			// if len(parts) > 2 && parts[2] != "settings" {
			// 	parts = append(parts[:2], append([]string{"builds"}, parts[2:]...)...)
			// }

			// prefix the URL with /repos so that it
			// can be effectively routed.
			parts = append([]string{"", "repos"}, parts...)

			// reconstruct the path
			r.URL.Path = strings.Join(parts, "/")
		}

		h.ServeHTTP(w, r)
	})
}
