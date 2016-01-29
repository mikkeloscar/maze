package repo

import (
	"os"
	"path"
	"testing"

	"github.com/mikkeloscar/gopkgbuild"
	"github.com/stretchr/testify/assert"
)

var (
	workdir, _ = os.Getwd()
	repo1      = Repo{
		Name: "repo1",
		Path: path.Join(workdir, "test_files"),
	}
	repo2 = Repo{
		Name: "repo2",
		Path: path.Join(workdir, "test_files"),
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

func TestAdd(t *testing.T) {
	pkgPaths := []string{
		"test_files/ca-certificates-20150402-1-any.pkg.tar.xz",
	}

	err := repo2.Add(pkgPaths)
	assert.NoError(t, err, "should not fail")

	// readd package
	err = repo2.Add(pkgPaths)
	assert.NoError(t, err, "should not fail")

	// clean
	err = os.Remove(repo2.DB())
	assert.NoError(t, err, "should not fail")
	err = os.Remove(repo2.DB() + ".old")
	assert.NoError(t, err, "should not fail")
	err = os.Remove(path.Join(repo2.Path, repo2.Name+".db"))
	assert.NoError(t, err, "should not fail")
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

func Testobsolete(t *testing.T) {
	whiteList := []string{
		"a",
		"b",
		"c",
		"d",
	}

	pkgMap := map[string]*pkgDep{
		"a": &pkgDep{
			depends: []string{"c"},
		},
		"b": &pkgDep{},
	}

	obsolete := repo1.obsolete(whiteList, pkgMap)
	assert.Len(t, obsolete, 1, "should have length 1")
	assert.Equal(t, "d", obsolete[0], "should be equal")
}

func TestinStrSlice(t *testing.T) {
	haystack := []string{"a", "b"}
	assert.True(t, inStrSlice("a", haystack), "should be true")
	assert.False(t, inStrSlice("c", haystack), "should be false")
}
