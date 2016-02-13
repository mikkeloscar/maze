package model

import "time"

type Repo struct {
	ID          int64      `json:"id"           meddler:"id,pk"`
	UserID      int64      `json:"-"            meddler:"user_id"`
	Owner       string     `json:"owner"	meddler:"owner"`
	Name        string     `json:"name"         meddler:"name"`
	Path        string     `json:"path"         meddler:"path"`
	SourceOwner string     `json:"source_owner" meddler:"source_owner"`
	SourceName  string     `json:"source_name"  meddler:"source_name"`
	Hash        string     `json:"-"            meddler:"hash"`
	LastCheck   *time.Time `json:"last_check"   meddler:"last_check"`
}
