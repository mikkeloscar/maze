package main

import (
	"flag"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/drone/drone/shared/envconfig"
	"github.com/gin-gonic/gin"
	"github.com/mikkeloscar/maze-repo/checker"
	"github.com/mikkeloscar/maze-repo/remote"
	"github.com/mikkeloscar/maze-repo/router"
	"github.com/mikkeloscar/maze-repo/router/middleware/context"
	"github.com/mikkeloscar/maze-repo/store/datastore"
)

var (
	envConf = flag.String("config", "env.conf", "")
	debug   = flag.Bool("d", false, "")
)

func main() {
	flag.Parse()

	if !*debug {
		// disbale gin debug mode
		gin.SetMode(gin.ReleaseMode)
	}

	env := envconfig.Load(*envConf)
	store_, err := datastore.Load(env)
	if err != nil {
		log.Errorf("failed to load datastore: %s", err)
		os.Exit(1)
	}
	remote_ := remote.Load(env)

	chck := checker.Checker{
		Remote: remote_,
		Store:  store_,
	}
	go chck.Run()

	router.Load(
		context.SetStore(store_),
		context.SetRemote(remote_),
	).Run(":8080")
}
