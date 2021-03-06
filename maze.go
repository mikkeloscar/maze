package main

import (
	"flag"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ianschenck/envflag"
	"github.com/mikkeloscar/maze/checker"
	"github.com/mikkeloscar/maze/remote"
	"github.com/mikkeloscar/maze/repo"
	"github.com/mikkeloscar/maze/router"
	"github.com/mikkeloscar/maze/router/middleware/context"
	"github.com/mikkeloscar/maze/store/datastore"
	log "github.com/sirupsen/logrus"
)

var (
	addr = envflag.String("SERVER_ADDR", ":8080", "")

	debug    = flag.Bool("d", false, "")
	check    = flag.Bool("check", false, "Enable automatic check of package updates.")
	stateTTL = 2 * time.Hour
)

func main() {
	flag.Parse()
	envflag.Parse()

	if !*debug {
		// disbale gin debug mode
		gin.SetMode(gin.ReleaseMode)
	}

	err := repo.LoadRepoStorage()
	if err != nil {
		log.Fatalf("repo storage error: %s", err)
	}

	log.Printf("using repo storage path: %s", repo.RepoStorage)

	ctxStore, err := datastore.Load()
	if err != nil {
		log.Fatalf("failed to load datastore: %s", err)
	}
	ctxRemote := remote.Load()

	middleware := []gin.HandlerFunc{
		context.SetStore(ctxStore),
		context.SetRemote(ctxRemote),
	}

	if *check {
		state := checker.NewState(stateTTL)

		chck := checker.Checker{
			Remote: ctxRemote,
			Store:  ctxStore,
			State:  state,
		}
		go chck.Run()

		middleware = append(middleware, context.SetState(state))
	}

	// setup the server and start listening
	handler := router.Load(middleware...)

	log.Fatal(http.ListenAndServe(*addr, handler))
}
