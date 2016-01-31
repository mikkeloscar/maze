package main

import "github.com/mikkeloscar/maze-repo/router"

func main() {
	// gin.SetMode(gin.ReleaseMode)
	router.Load().Run(":8080")
}
