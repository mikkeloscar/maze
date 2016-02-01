package remote

import (
	"net/http"

	"github.com/mikkeloscar/maze-repo/model"
)

type Remote interface {
	// Login authenticates the session and returns the remoter user
	// details.
	Login(res http.ResponseWriter, req *http.Request) (*model.User, error)

	// EmptyCommit creates/adds a new empty commit to a branch of a repo.
	// if srcBranch and dstBranch are different then the commit will
	// include the state of srcbranch effectively rebasing dstBranch onto
	// srcBranch.
	EmptyCommit(u *model.User, repo, srcBranch, dstBranch, msg string) error

	// SetupBranch sets up a new branch based on srcBranch. If the branch
	// already exists nothing happens.
	SetupBranch(u *model.User, repo, srcBranch, dstBranch string) error
}
