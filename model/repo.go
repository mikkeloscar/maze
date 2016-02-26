package model

import "time"

type Repo struct {
	ID           int64     `json:"id"            meddler:"id,pk"`
	UserID       int64     `json:"-"             meddler:"user_id"`
	Private      bool      `json:"private"       meddler:"private"`
	Owner        string    `json:"owner"         meddler:"owner"`
	Name         string    `json:"name"          meddler:"name"`
	SourceOwner  string    `json:"source_owner"  meddler:"source_owner"`
	SourceName   string    `json:"source_name"   meddler:"source_name"`
	SourceBranch string    `json:"source_branch" meddler:"source_branch"`
	BuildBranch  string    `json:"build_branch"  meddler:"build_branch"`
	Hash         string    `json:"-"             meddler:"hash"`
	LastCheck    time.Time `json:"last_check"    meddler:"last_check,utctime"`
}
