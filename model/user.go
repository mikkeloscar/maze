package model

type User struct {
	ID    int64  `json:"id" meddler:"user_id,pk"`
	Login string `json:"login" meddler:"user_login"`
	Token string `json:"-" meddler:"user_token"`
}
