package repo

import (
	"testing"

	"github.com/mikkeloscar/gopkgbuild"
	"github.com/stretchr/testify/assert"
)

var (
	repo1 = Repo{
		Name: "repo1",
		Path: "test_files",
	}
	repo2 = Repo{
		Name: "repo2",
		Path: "test_files",
	}
)

// Test splitting name and version string.
func TestSplitNameVersion(t *testing.T) {
	name, version := splitNameVersion("ca-certificates-20150402-1/")
	assert.Equal(t, "ca-certificates", name, "should be equal")
	assert.Equal(t, "20150402-1", version, "should be equal")

	name, version = splitNameVersion("ca-certificates-2:20150402-1/")
	assert.Equal(t, "ca-certificates", name, "should be equal")
	assert.Equal(t, "2:20150402-1", version, "should be equal")

	name, version = splitNameVersion("zlib-1.2.8-4/")
	assert.Equal(t, "zlib", name, "should be equal")
	assert.Equal(t, "1.2.8-4", version, "should be equal")
}

// Test IsNew.
func TestIsNew(t *testing.T) {
	pkg := "ca-certificates"

	// Check if existing package is new
	version, _ := pkgbuild.NewCompleteVersion("20150402-1")
	new, err := repo1.IsNew(pkg, *version)
	assert.NoError(t, err, "should not fail")
	assert.False(t, new, "should be false")

	// Check if new package is new
	version, _ = pkgbuild.NewCompleteVersion("20150402-2")
	new, err = repo1.IsNew(pkg, *version)
	assert.NoError(t, err, "should not fail")
	assert.True(t, new, "should be true")

	// Check if old package is new
	version, _ = pkgbuild.NewCompleteVersion("20150401-1")
	new, err = repo1.IsNew(pkg, *version)
	assert.NoError(t, err, "should not fail")
	assert.False(t, new, "should be false")

	// Check if existing package is new (repo is empty)
	version, _ = pkgbuild.NewCompleteVersion("20150402-1")
	new, err = repo2.IsNew(pkg, *version)
	assert.NoError(t, err, "should not fail")
	assert.True(t, new, "should be true")
}
