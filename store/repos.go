package store

import (
	"github.com/mikkeloscar/maze-repo/model"
	"golang.org/x/net/context"
)

type RepoStore interface {
	// Get gets a repo by unique ID.
	Get(int64) (*model.Repo, error)

	// GetByName gets a repo by owner and name.
	GetByName(string, string) (*model.Repo, error)

	// Get all repos.
	GetRepoList() ([]*model.Repo, error)

	// Create creates a new repository.
	Create(*model.Repo) error

	// Update updates a repository.
	Update(*model.Repo) error

	// Delete deletes a user repository.
	Delete(*model.Repo) error
}

func GetRepo(c context.Context, id int64) (*model.Repo, error) {
	return FromContext(c).Repos().Get(id)
}

func GetRepoByOwnerName(c context.Context, owner, name string) (*model.Repo, error) {
	return FromContext(c).Repos().GetByName(owner, name)
}

func CreateRepo(c context.Context, repo *model.Repo) error {
	return FromContext(c).Repos().Create(repo)
}

func UpdateRepo(c context.Context, repo *model.Repo) error {
	return FromContext(c).Repos().Update(repo)
}

func DeleteRepo(c context.Context, repo *model.Repo) error {
	return FromContext(c).Repos().Delete(repo)
}
