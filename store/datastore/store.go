package datastore

import (
	"database/sql"
	"os"
	"path"

	log "github.com/Sirupsen/logrus"
	"github.com/ianschenck/envflag"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mikkeloscar/maze/store"
	"github.com/rubenv/sql-migrate"
	"github.com/russross/meddler"
)

var (
	driver = envflag.String("DATABASE_DRIVER", "sqlite3", "")
	config = envflag.String("DATABASE_CONFIG", "maze.sqlite", "")
)

func Load() (store.Store, error) {
	return New(*driver, *config)
}

func New(driver, config string) (store.Store, error) {
	db, err := Open(driver, config)
	if err != nil {
		return nil, err
	}

	return store.New(
		driver,
		&userStore{db},
		&repoStore{db},
	), nil
}

func Open(driver, config string) (*sql.DB, error) {
	db, err := sql.Open(driver, config)
	if err != nil {
		return nil, err
	}

	switch driver {
	case "sqlite3":
		meddler.Default = meddler.SQLite
	case "mysql":
		meddler.Default = meddler.MySQL
	}

	// Migrate db
	err = migrateDB(driver, db)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func migrateDB(driver string, db *sql.DB) error {
	log.Printf("migrating db")
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	migrations := &migrate.FileMigrationSource{
		Dir: path.Join(wd, "store", "migration", driver),
	}
	_, err = migrate.Exec(db, driver, migrations, migrate.Up)
	return err
}
