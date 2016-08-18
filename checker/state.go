package checker

import (
	"sync"
	"time"
)

// State defines a table of the packages for which an update/check request was
// made recently.
type State struct {
	table  map[string]map[string]time.Time
	rwLock *sync.RWMutex
	ttl    time.Duration
}

// NewState initializes a new state with the specified ttl.
func NewState(ttl time.Duration) *State {
	return &State{
		make(map[string]map[string]time.Time),
		new(sync.RWMutex),
		ttl,
	}
}

// Add adds a package to the state table for the given repo.
func (s *State) Add(pkg, owner, repo string) {
	s.rwLock.Lock()
	defer s.rwLock.Unlock()

	repo = owner + "/" + repo

	if pkgs, ok := s.table[repo]; ok {
		pkgs[pkg] = time.Now().UTC()
	} else {
		pkgs = make(map[string]time.Time)
		pkgs[pkg] = time.Now().UTC()
		s.table[repo] = pkgs
	}
}

// ClearPkg clears a package from the state table.
func (s *State) ClearPkg(pkg, owner, repo string) {
	s.rwLock.Lock()
	defer s.rwLock.Unlock()

	repo = owner + "/" + repo

	if pkgs, ok := s.table[repo]; ok {
		delete(pkgs, pkg)
		if len(pkgs) == 0 {
			delete(s.table, repo)
		}
	}
}

// IsActive returns true (and add time) if a check/update request has recently
// been made for this package.
func (s *State) IsActive(pkg, owner, repo string) (bool, *time.Time) {
	s.rwLock.RLock()
	defer s.rwLock.RUnlock()

	repo = owner + "/" + repo

	if pkgs, ok := s.table[repo]; ok {
		if t, ok := pkgs[pkg]; ok {
			now := time.Now().UTC()
			ttl := t.Add(s.ttl)
			return ttl.After(now), &t
		}
	}

	return false, nil
}

// ClearExpired clears all expired entries from the state table.
func (s *State) ClearExpired() {
	s.rwLock.Lock()
	defer s.rwLock.Unlock()

	for repo, pkgs := range s.table {
		for pkg, t := range pkgs {
			now := time.Now().UTC()
			ttl := t.Add(s.ttl)
			if ttl.Before(now) {
				delete(pkgs, pkg)
				if len(pkgs) == 0 {
					delete(s.table, repo)
				}
			}
		}
	}
}
