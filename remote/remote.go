package remote

import (
	"net/http"

	"github.com/ianschenck/envflag"
	"github.com/mikkeloscar/maze/common/pkgconfig"
	"github.com/mikkeloscar/maze/model"
	"github.com/mikkeloscar/maze/remote/github"
)

var (
	client = envflag.String("CLIENT", "", "")
	secret = envflag.String("SECRET", "", "")
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

func Load() Remote {
	return github.Load(*client, *secret)
}
