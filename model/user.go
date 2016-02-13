package model

type User struct {
	ID    int64  `json:"id"    meddler:"id,pk"`
	Login string `json:"login" meddler:"login"`
	Token string `json:"-"     meddler:"token"`
	Admin bool   `json:"admin" meddler:"admin"`
	Hash  string `json:"-"     meddler:"hash"`
}
