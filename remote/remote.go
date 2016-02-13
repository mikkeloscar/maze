package remote

import (
	"net/http"

	"github.com/drone/drone/shared/envconfig"
	"github.com/mikkeloscar/maze-repo/common/pkgconfig"
	"github.com/mikkeloscar/maze-repo/model"
	"github.com/mikkeloscar/maze-repo/remote/github"
)

type Remote interface {
	// Login authenticates the session and returns the remoter user
	// details.
	Login(res http.ResponseWriter, req *http.Request) (*model.User, error)

	// Repo fetches the named repository from the remote system.
	Repo(u *model.User, owner, name string) (*model.Repo, error)

	// Perm fetches the named repository permissions from the remote system
	// for the specified user.
	Perm(u *model.User, owner, name string) (*model.Perm, error)

	// EmptyCommit creates/adds a new empty commit to a branch of a repo.
	// if srcBranch and dstBranch are different then the commit will
	// include the state of srcbranch effectively rebasing dstBranch onto
	// srcBranch.
	EmptyCommit(u *model.User, owner, repo, srcBranch, dstBranch, msg string) error

	// SetupBranch sets up a new branch based on srcBranch. If the branch
	// already exists nothing happens.
	SetupBranch(u *model.User, owner, repo, srcBranch, dstBranch string) error

	// GetConfig gets and parses the package.yml config file.
	GetConfig(u *model.User, owner, repo, path string) (*pkgconfig.PkgConfig, error)
}

func Load(env envconfig.Env) Remote {
	return github.Load(env)
}
