package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStrContains(t *testing.T) {
	haystack := []string{"a", "b"}
	assert.True(t, StrContains("a", haystack), "should be true")
	assert.False(t, StrContains("c", haystack), "should be false")
}

func TestIsDevel(t *testing.T) {
	assert.True(t, IsDevel("maze-git"), "should be true")
	assert.True(t, IsDevel("maze-hg"), "should be true")
	assert.True(t, IsDevel("maze-bzr"), "should be true")
	assert.True(t, IsDevel("maze-svn"), "should be true")
	assert.False(t, IsDevel("maze-foo"), "should be false")
	assert.False(t, IsDevel("maze"), "should be false")
}
