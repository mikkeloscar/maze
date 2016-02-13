package store

type Store interface {
	Users() UserStore
	Repos() RepoStore
}

type store struct {
	name  string
	users UserStore
	repos RepoStore
}

func (s *store) Users() UserStore {
	return s.users
}

func (s *store) Repos() RepoStore {
	return s.repos
}

func New(name string, users UserStore, repos RepoStore) Store {
	return &store{
		name,
		users,
		repos,
	}
}
