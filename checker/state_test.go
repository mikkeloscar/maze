package checker

import (
	"testing"
	"time"
)

// TestAddPkg tests adding a package to the state table.
func TestAddPkg(t *testing.T) {
	s := NewState(time.Duration(1 * time.Second))
	pkg := "abc"
	owner := "owner"
	repo := "xyz"
	s.Add(pkg, owner, repo)

	r, ok := s.table[owner+"/"+repo]
	if !ok {
		t.Errorf("Should find repo in state table")
	}

	_, ok = r[pkg]
	if !ok {
		t.Errorf("Should find pkg in repo table")
	}
}

// TestClearPkg tests clearing a package form the state table.
func TestClearPkg(t *testing.T) {
	s := NewState(time.Duration(1 * time.Second))
	pkg := "abc"
	owner := "owner"
	repo := "xyz"
	s.Add(pkg, owner, repo)

	s.ClearPkg(pkg, owner, repo)

	_, ok := s.table[owner+"/"+repo]
	if ok {
		t.Errorf("Should not find repo in state table")
	}

	// nothing should happen if the pkg has already been removed.
	s.ClearPkg(pkg, owner, repo)
}

// TestIsActive tests the IsActive method.
func TestIsActive(t *testing.T) {
	s := NewState(time.Duration(10 * time.Second))
	pkg := "abc"
	owner := "owner"
	repo := "xyz"
	s.Add(pkg, owner, repo)

	r, ok := s.table[owner+"/"+repo]
	if !ok {
		t.Errorf("Should find repo in state table")
	}

	// decrease time artificially
	r[pkg] = time.Now().UTC().Add(time.Duration(-20 * time.Minute))

	active, _ := s.IsActive(pkg, owner, repo)
	if active {
		t.Errorf("Should not be active anymore")
	}

	// increase time artificially
	r[pkg] = time.Now().UTC().Add(time.Duration(20 * time.Minute))
	active, _ = s.IsActive(pkg, owner, repo)
	if !active {
		t.Errorf("Should still be active")
	}

}

// TestClearExpired tests that expired entries are removed from the table.
func TestClearExpired(t *testing.T) {
	s := NewState(time.Duration(1 * time.Minute))
	pkg := "abc"
	owner := "owner"
	repo := "xyz"
	s.Add(pkg, owner, repo)
	s.Add("abcd", "owner", "xyzz")

	before := len(s.table)

	r, ok := s.table[owner+"/"+repo]
	if !ok {
		t.Errorf("Should find repo in state table")
	}
	// decrease time artificially
	r[pkg] = time.Now().UTC().Add(time.Duration(-10 * time.Minute))

	s.ClearExpired()

	if len(s.table) != before-1 {
		t.Errorf("Should find %d entry found %d", before-1, len(s.table))
	}
}
