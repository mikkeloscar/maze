package model

type Perm struct {
	Read  bool `json:"read"`
	Write bool `json:"write"`
	Admin bool `json:"admin"`
}
