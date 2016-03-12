package repo

import (
	"os"
	"os/exec"
	"path"
	"testing"

	"github.com/mikkeloscar/gopkgbuild"
	"github.com/mikkeloscar/maze/model"
	"github.com/stretchr/testify/assert"
)

var (
	workdir, _  = os.Getwd()
	repoStorage = path.Join(workdir, "test_files")
	repo1       = NewRepo(&model.Repo{Name: "repo1"}, repoStorage)
	repo2       = NewRepo(&model.Repo{Name: "repo2"}, repoStorage)
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

// Test splitting name and version filename.
func TestSplitFileNameVersion(t *testing.T) {
	name, version, arch, err := splitFileNameVersion("ca-certificates-20150402-1-any.pkg.tar.xz")
	assert.NoError(t, err, "should not fail")
	assert.Equal(t, "ca-certificates", name, "should be equal")
	assert.Equal(t, "any", arch, "should be equal")
	assert.Equal(t, "20150402-1", version, "should be equal")

	name, version, arch, err = splitFileNameVersion("ca-certificates-2:20150402-1-any.pkg.tar.xz")
	assert.NoError(t, err, "should not fail")
	assert.Equal(t, "ca-certificates", name, "should be equal")
	assert.Equal(t, "any", arch, "should be equal")
	assert.Equal(t, "2:20150402-1", version, "should be equal")

	name, version, arch, err = splitFileNameVersion("zlib-1.2.8-4-x86_64.pkg.tar.xz")
	assert.NoError(t, err, "should not fail")
	assert.Equal(t, "zlib", name, "should be equal")
	assert.Equal(t, "x86_64", arch, "should be equal")
	assert.Equal(t, "1.2.8-4", version, "should be equal")

	_, _, _, err = splitFileNameVersion("zlib-1.2.8-any.pkg.tar.xz")
	assert.Error(t, err, "should fail")

	_, _, _, err = splitFileNameVersion("zlib1.2.8-4-any.pkg.tar.xz")
	assert.Error(t, err, "should fail")

	_, _, _, err = splitFileNameVersion("zlib-1.2.8-4-any.pkg.tar.xz")
	assert.NoError(t, err, "should not fail")

	name, version, arch, err = splitFileNameVersion("pulseaudio-raop2-8.0-1-x86_64.pkg.tar.xz")
	assert.NoError(t, err, "should not fail")
	assert.Equal(t, "pulseaudio-raop2", name, "should be equal")
	assert.Equal(t, "x86_64", arch, "should be equal")
	assert.Equal(t, "8.0-1", version, "should be equal")
}

func TestAdd(t *testing.T) {
	err := repo2.InitDir()
	assert.NoError(t, err, "should not fail")

	pkgPaths := []string{
		"test_files/repo2/x86_64/ca-certificates-20150402-1-any.pkg.tar.xz",
	}

	cmd := exec.Command(
		"cp",
		"test_files/repo1/x86_64/ca-certificates-20150402-1-any.pkg.tar.xz",
		"test_files/repo2/x86_64/ca-certificates-20150402-1-any.pkg.tar.xz")
	err = cmd.Run()
	assert.NoError(t, err, "should not fail")

	err = repo2.Add(pkgPaths)
	assert.NoError(t, err, "should not fail")

	err = repo2.Add([]string{})
	assert.NoError(t, err, "should not fail")

	// readd package
	err = repo2.Add(pkgPaths)
	assert.NoError(t, err, "should not fail")

	// clean
	err = repo2.ClearPath()
	assert.NoError(t, err, "should not fail")
}

// Test Remove.
func TestRemove(t *testing.T) {
	// setup test repo
	err := repo2.InitDir()
	assert.NoError(t, err, "should not fail")

	pkgPaths := []string{
		"test_files/repo2/x86_64/ca-certificates-20150402-1-any.pkg.tar.xz",
	}

	cmd := exec.Command(
		"cp",
		"test_files/repo1/x86_64/ca-certificates-20150402-1-any.pkg.tar.xz",
		"test_files/repo2/x86_64/ca-certificates-20150402-1-any.pkg.tar.xz")
	err = cmd.Run()
	assert.NoError(t, err, "should not fail")

	err = repo2.Add(pkgPaths)
	assert.NoError(t, err, "should not fail")

	// remove package
	err = repo2.Remove([]string{"ca-certificates"}, "x86_64")
	assert.NoError(t, err, "should not fail")

	// clean
	err = repo2.ClearPath()
	assert.NoError(t, err, "should not fail")
}

// Test IsNew.
func TestIsNew(t *testing.T) {
	pkg := "ca-certificates"

	// Check if existing package is new
	version, _ := pkgbuild.NewCompleteVersion("20150402-1")
	new, err := repo1.IsNew(pkg, "any", *version)
	assert.NoError(t, err, "should not fail")
	assert.False(t, new, "should be false")

	// Check if new package is new
	version, _ = pkgbuild.NewCompleteVersion("20150402-2")
	new, err = repo1.IsNew(pkg, "any", *version)
	assert.NoError(t, err, "should not fail")
	assert.True(t, new, "should be true")

	// Check if old package is new
	version, _ = pkgbuild.NewCompleteVersion("20150401-1")
	new, err = repo1.IsNew(pkg, "any", *version)
	assert.NoError(t, err, "should not fail")
	assert.False(t, new, "should be false")

	// Check if existing package is new (repo is empty)
	version, _ = pkgbuild.NewCompleteVersion("20150402-1")
	new, err = repo2.IsNew(pkg, "any", *version)
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

func TestPackages(t *testing.T) {
	pkgs, err := repo1.Packages("x86_64", false)
	assert.NoError(t, err, "should not fail")
	assert.Len(t, pkgs, 1, "should have length 1")

	pkgs, err = repo1.Packages("x86_64", true)
	assert.NoError(t, err, "should not fail")
	assert.Len(t, pkgs, 1, "should have length 1")
}

func TestPackage(t *testing.T) {
	pkg, err := repo1.Package("ca-certificates", "x86_64", false)
	assert.NoError(t, err, "should not fail")
	assert.NotNil(t, pkg, "should not be nil")
	assert.Equal(t, pkg.Name, "ca-certificates", "should be equal")

	// package that doesn't exist
	pkg, err = repo1.Package("ca-certificates-foo", "x86_64", false)
	assert.NoError(t, err, "should not fail")
	assert.Nil(t, pkg, "should be nil")
}

func TestValidRepoName(t *testing.T) {
	assert.True(t, ValidRepoName("test"), "should be true")
	assert.True(t, ValidRepoName("test123"), "should be true")
	assert.True(t, ValidRepoName("test@"), "should be true")
	assert.True(t, ValidRepoName("test."), "should be true")
	assert.True(t, ValidRepoName("test_"), "should be true")
	assert.True(t, ValidRepoName("test+"), "should be true")
	assert.True(t, ValidRepoName("test-"), "should be true")
	assert.True(t, ValidRepoName("@test-"), "should be true")
	assert.False(t, ValidRepoName("-test"), "should be false")
	assert.False(t, ValidRepoName("test="), "should be false")
	assert.False(t, ValidRepoName("te st"), "should be false")
}
