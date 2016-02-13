package datastore

import (
	"database/sql"

	"github.com/mikkeloscar/maze-repo/model"
	"github.com/russross/meddler"
)

type repoStore struct {
	*sql.DB
}

func (db *repoStore) Get(ID int64) (*model.Repo, error) {
	repo := new(model.Repo)
	err := meddler.Load(db, repoTable, repo, ID)
	return repo, err
}

func (db *repoStore) GetByName(owner, name string) (*model.Repo, error) {
	repo := new(model.Repo)
	err := meddler.QueryRow(db, repo, repoNameQuery, owner, name)
	return repo, err
}

func (db *repoStore) GetRepoList() ([]*model.Repo, error) {
	var repos []*model.Repo
	err := meddler.QueryAll(db, &repos, repoListQuery)
	return repos, err
}

func (db *repoStore) Create(repo *model.Repo) error {
	return meddler.Insert(db, repoTable, repo)
}

func (db *repoStore) Update(repo *model.Repo) error {
	return meddler.Update(db, repoTable, repo)
}

func (db *repoStore) Delete(repo *model.Repo) error {
	_, err := db.Exec(repoDeleteQuery, repo.ID)
	return err
}

const repoTable = "repos"

const repoNameQuery = `
SELECT *
FROM repos
WHERE owner = ? AND name = ? 
LIMIT 1
`

const repoListQuery = `
SELECT *
FROM repos
ORDER BY last_check
`

const repoDeleteQuery = `
DELETE FROM repos
WHERE id = ?
`
