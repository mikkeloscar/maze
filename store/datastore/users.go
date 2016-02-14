package datastore

import (
	"database/sql"

	"github.com/mikkeloscar/maze/model"
	"github.com/russross/meddler"
)

type userStore struct {
	*sql.DB
}

func (db *userStore) Get(id int64) (*model.User, error) {
	user := new(model.User)
	err := meddler.Load(db, userTable, user, id)
	return user, err
}

func (db *userStore) GetLogin(login string) (*model.User, error) {
	user := new(model.User)
	err := meddler.QueryRow(db, user, userLoginQuery, login)
	return user, err
}

func (db *userStore) Count() (int, error) {
	var count int
	err := db.QueryRow(userCountQuery).Scan(&count)
	return count, err
}

func (db *userStore) Create(user *model.User) error {
	return meddler.Insert(db, userTable, user)
}

func (db *userStore) Update(user *model.User) error {
	return meddler.Update(db, userTable, user)
}

func (db *userStore) Delete(user *model.User) error {
	_, err := db.Exec(userDeleteQuery, user.ID)
	return err
}

const userTable = "users"

const userLoginQuery = `
SELECT id, login, token, admin, hash
FROM users
WHERE login=?
LIMIT 1
`

const userCountQuery = `
SELECT count(1)
FROM users
`

const userDeleteQuery = `
DELETE FROM users
WHERE id=?
`
