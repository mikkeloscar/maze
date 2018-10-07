package aur

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetDeps(t *testing.T) {
	pkgs := []string{
		"virtualbox-guest-modules-mainline",
		"virtualbox-host-modules-mainline",
		"linux-mainline",
		"bbswitch-mainline",
		"pacaur",
		"cower",
	}

	deps := make(map[string]*depNode)

	err := getDeps(pkgs, nil, deps)
	assert.NoError(t, err, "should not fail")

	groups := groupDeps(deps)
	assert.Len(t, groups, 2, "should have len 2")

	pkgs = []string{
		"sway-git",
		"wlc-git",
	}

	deps = make(map[string]*depNode)

	err = getDeps(pkgs, nil, deps)
	assert.NoError(t, err, "should not fail")

	groups = groupDeps(deps)
	assert.Len(t, groups, 2, "should have len 2")
}
