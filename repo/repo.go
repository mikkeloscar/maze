package repo

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"

	"github.com/mikkeloscar/gopkgbuild"
)

var pkgPatt = regexp.MustCompile(`[a-z]+[a-z\-]+[a-z]+-(\d+:)?[\da-z\.]+-\d+-(i686|x86_64|any).pkg.tar.xz(.sig)?`)
var repos = map[string]*Repo{
	"test": &Repo{
		Name: "test",
		Path: "/home/moscar/projects/go/src/github.com/mikkeloscar/maze-repo/repo/test_db",
	},
}

func GetByName(name string) *Repo {
	v, _ := repos[name]
	return v
}

// Repo is a wrapper around the arch tools 'repo-add' and 'repo-remove'.
type Repo struct {
	Name  string
	Path  string
	Files bool
}

// DB returns path to db archive.
func (r *Repo) DB() string {
	return path.Join(r.Path, fmt.Sprintf("%s.db.tar.gz", r.Name))
}

// Add adds a list of packages to a repo db, moving the package files to
// the repo db directory if needed.
func (r *Repo) Add(pkgPaths []string) error {
	for i, pkg := range pkgPaths {
		pkgPathDir, pkgPathBase := path.Split(pkg)

		if r.Path != pkgPathDir {
			// move pkg to repo path.
			newPath := path.Join(r.Path, pkgPathBase)
			err := os.Rename(pkg, newPath)
			if err != nil {
				return err
			}
			pkgPaths[i] = newPath
		}
	}

	args := []string{"-R", r.DB()}
	args = append(args, pkgPaths...)

	cmd := exec.Command("repo-add", args...)
	cmd.Dir = r.Path

	return cmd.Run()
}

// Remove removes a list of packages from the repo db.
func (r *Repo) Remove(pkgs []string) error {
	args := []string{"-R", r.DB()}
	args = append(args, pkgs...)

	cmd := exec.Command("repo-remove", args...)
	cmd.Dir = r.Path

	return cmd.Run()
}

func (r *Repo) IsNewFilename(file string) (bool, error) {
	name, version, err := splitFileNameVersion(file)
	if err != nil {
		return false, err
	}

	ver, err := pkgbuild.NewCompleteVersion(version)
	if err != nil {
		return false, err
	}

	return r.IsNew(name, *ver)
}

// IsNew returns true if pkg is a newer version than what's in the repo.
// If the package is not found in the repo, it will be marked as new.
func (r *Repo) IsNew(name string, version pkgbuild.CompleteVersion) (bool, error) {
	f, err := os.Open(r.DB())
	if err != nil {
		if os.IsNotExist(err) {
			return true, nil
		}
		return false, err
	}
	defer f.Close()

	gzf, err := gzip.NewReader(f)
	if err != nil {
		return false, err
	}

	tarR := tar.NewReader(gzf)

	for {
		header, err := tarR.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			return false, err
		}

		switch header.Typeflag {
		case tar.TypeDir:
			n, v := splitNameVersion(header.Name)
			if n == name {
				if version.Newer(v) {
					return true, nil
				}
				return false, nil
			}
		case tar.TypeReg:
			continue
		}
	}

	return true, nil
}

// Obsolete returns a list of obsolete packages based on the input packages.
// A package is considered obsolete if it's not in the input list and not a
// dependency of one of the input packages.
func (r *Repo) Obsolete(pkgs []string) ([]string, error) {
	pkgMap, err := r.readPkgMap()
	if err != nil {
		return nil, err
	}

	return r.obsolete(pkgs, pkgMap), nil
}

func (r *Repo) obsolete(pkgs []string, pkgMap map[string]*pkgDep) []string {
	var obsolete map[string]struct{}

	for n := range pkgMap {
		if !inStrSlice(n, pkgs) {
			obsolete[n] = struct{}{}
		}
	}

	for _, p := range pkgs {
		if deps, ok := pkgMap[p]; ok {
			for _, dep := range deps.depends {
				if _, ok := obsolete[dep]; ok {
					delete(obsolete, dep)
					break
				}
			}
		}
	}

	obsol := make([]string, 0, len(obsolete))
	for pkg := range obsolete {
		obsol = append(obsol, pkg)
	}

	return obsol
}

func inStrSlice(needle string, haystack []string) bool {
	for _, item := range haystack {
		if item == needle {
			return true
		}
	}
	return false
}

// TODO: what is a needed dep?
type pkgDep struct {
	makedepends []string
	depends     []string
	optdepends  []string
}

func (r *Repo) readPkgMap() (map[string]*pkgDep, error) {
	pkgMap := make(map[string]*pkgDep)

	f, err := os.Open(r.DB())
	if err != nil {
		return nil, err
	}
	defer f.Close()

	gzf, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}

	tarR := tar.NewReader(gzf)

	for {
		header, err := tarR.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
		}

		switch header.Typeflag {
		case tar.TypeReg:
			if strings.HasSuffix(header.Name, "/depends") {
				rdr := bufio.NewReader(tarR)
				pkg := &pkgDep{}
				var curr []string
				eof := false
				for {
					line, err := rdr.ReadString('\n')
					if err != nil {
						if err != io.EOF {
							return nil, err
						}

						eof = true
					}

					switch line {
					case "%DEPENDS%":
						curr = pkg.depends
					case "%MAKEDEPENDS%":
						curr = pkg.makedepends
					case "%OPTDEPENDS%":
						curr = pkg.optdepends
					case "":
						if eof {
							break
						}
					default:
						curr = append(curr, line)
					}
				}
				pkgMap[path.Dir(header.Name)] = pkg
				fmt.Println(pkg)
			}
		}
	}

	return pkgMap, nil
}

// turn "zlib-1.2.8-4/" into ("zlib", "1.2.8-4").
func splitNameVersion(str string) (string, string) {
	chars := strings.Split(str[:len(str)-1], "-")
	name := chars[:len(chars)-2]
	version := chars[len(chars)-2:]

	return strings.Join(name, "-"), strings.Join(version, "-")
}

// turn "zlib-1.2.8-4-x86_64.pkg.tar.xz" into ("zlib", "1.2.8-4").
func splitFileNameVersion(file string) (string, string, error) {
	if pkgPatt.MatchString(file) {
		sections := strings.Split(file, "-")
		if len(sections) > 3 {
			name := sections[:len(sections)-3]
			version := sections[len(sections)-3 : len(sections)-1]
			return strings.Join(name, "-"), strings.Join(version, "-"), nil
		}
	}

	return "", "", fmt.Errorf("invalid package filename")
}
