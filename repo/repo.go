package repo

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/mikkeloscar/gopkgbuild"
	"github.com/mikkeloscar/maze/model"
)

var pkgPatt = regexp.MustCompile(`[a-z]+[a-z\-]+[a-z]+-(\d+:)?[\da-z\.]+-\d+-(i686|x86_64|any).pkg.tar.xz(.sig)?`)

// Repo is a wrapper around the arch tools 'repo-add' and 'repo-remove'.
type Repo struct {
	*model.Repo
	basePath string
}

func NewRepo(r *model.Repo) *Repo {
	return newRepo(r, RepoStorage)
}

func newRepo(r *model.Repo, basePath string) *Repo {
	return &Repo{r, basePath}
}

func (r *Repo) InitDir() error {
	return os.MkdirAll(r.Path(), 0755)
}

func (r *Repo) ClearPath() error {
	return os.RemoveAll(r.Path())
}

func (r *Repo) Path() string {
	return path.Join(r.basePath, r.Owner, r.Name)
}

// DB returns path to db archive.
func (r *Repo) DB() string {
	return path.Join(r.Path(), fmt.Sprintf("%s.db.tar.gz", r.Name))
}

// FilesDB returns path to files db archive.
func (r *Repo) FilesDB() string {
	return path.Join(r.Path(), fmt.Sprintf("%s.files.tar.gz", r.Name))
}

// Add adds a list of packages to a repo db, moving the package files to
// the repo db directory if needed.
func (r *Repo) Add(pkgPaths []string) error {
	if len(pkgPaths) == 0 {
		return nil
	}

	for i, pkg := range pkgPaths {
		pkgPathDir, pkgPathBase := path.Split(pkg)

		if r.Path() != pkgPathDir {
			// move pkg to repo path.
			newPath := path.Join(r.Path(), pkgPathBase)
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
	cmd.Dir = r.Path()

	return cmd.Run()
}

// Remove removes a list of packages from the repo db.
func (r *Repo) Remove(pkgs []string) error {
	args := []string{"-R", r.DB()}
	args = append(args, pkgs...)

	cmd := exec.Command("repo-remove", args...)
	cmd.Dir = r.Path()

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

func parsePackage(tarRdr io.Reader, pkg *model.Package) error {
	rdr := bufio.NewReader(tarRdr)
	var curr *string
	var currTime *time.Time
	var currSlice *[]string
	for {
		line, err := rdr.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				return err
			}

			break
		}

		line = line[:len(line)-1]

		switch line {
		case `%FILENAME%`:
			curr = &pkg.FileName
		case `%NAME%`:
			curr = &pkg.Name
		case `%BASE%`:
			curr = &pkg.Base
		case `%VERSION%`:
			curr = &pkg.Version
		case `%DESC%`:
			curr = &pkg.Desc
		case `%CSIZE%`:
			curr = &pkg.CSize
		case `%ISIZE%`:
			curr = &pkg.ISize
		case `%MD5SUM%`:
			curr = &pkg.MD5Sum
		case `%SHA256SUM%`:
			curr = &pkg.SHA256Sum
		case `%URL%`:
			curr = &pkg.URL
		case `%LICENSE%`:
			curr = &pkg.License
		case `%ARCH%`:
			curr = &pkg.Arch
		case `%BUILDDATE%`:
			currTime = &pkg.BuildDate
		case `%PACKAGER%`:
			curr = &pkg.Packager
		case `%DEPENDS%`:
			currSlice = &pkg.Depends
		case `%MAKEDEPENDS%`:
			currSlice = &pkg.MakeDepends
		case `%OPTDEPENDS%`:
			currSlice = &pkg.OptDepends
		case ``:
			curr = nil
			currTime = nil
			currSlice = nil
		default:
			if curr != nil {
				*curr = line
			}

			if currTime != nil {
				i, err := strconv.ParseInt(line, 10, 64)
				if err != nil {
					return err
				}
				*currTime = time.Unix(i, 0)
			}

			if currSlice != nil {
				*currSlice = append(*currSlice, line)
			}
		}
	}

	return nil
}

// Package returns a named package from the repo.
func (r *Repo) Package(name string, files bool) (*model.Package, error) {
	f, err := os.Open(r.FilesDB())
	if err != nil {
		return nil, err
	}
	defer f.Close()

	gzf, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}

	tarR := tar.NewReader(gzf)

	found := false
	foundFiles := false
	pkg := &model.Package{}

	for {
		if found && (foundFiles || !files) {
			break
		}

		header, err := tarR.Next()
		if err != nil {
			if err != io.EOF {
				return nil, err
			}

			break
		}

		switch header.Typeflag {
		case tar.TypeReg:
			pkgName, _ := splitNameVersion(header.Name)

			if pkgName == name {
				if strings.HasSuffix(header.Name, "/desc") {
					err := parsePackage(tarR, pkg)
					if err != nil {
						return nil, err
					}
					found = true
				}
				if strings.HasSuffix(header.Name, "/files") && files {
					content, err := ioutil.ReadAll(tarR)
					if err != nil {
						return nil, err
					}

					pkg.Files = strings.Split(string(content), "\n")
					pkg.Files = pkg.Files[1:]
					if pkg.Files[len(pkg.Files)-1] == "" {
						pkg.Files = pkg.Files[:len(pkg.Files)-1]
					}
					foundFiles = true
				}
			}
		}
	}

	return pkg, nil
}

// Packages returns a list of all packages in the repo.
func (r *Repo) Packages(files bool) ([]*model.Package, error) {
	var pkgs []*model.Package

	f, err := os.Open(r.FilesDB())
	if err != nil {
		return nil, err
	}
	defer f.Close()

	gzf, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}

	tarR := tar.NewReader(gzf)

	var pkg *model.Package

	for {
		header, err := tarR.Next()
		if err != nil {
			if err != io.EOF {
				return nil, err
			}

			break
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if pkg != nil {
				pkgs = append(pkgs, pkg)
			}

			pkg = &model.Package{}
		case tar.TypeReg:
			if strings.HasSuffix(header.Name, "/desc") {
				err := parsePackage(tarR, pkg)
				if err != nil {
					return nil, err
				}
			}
			if strings.HasSuffix(header.Name, "/files") && files {
				content, err := ioutil.ReadAll(tarR)
				if err != nil {
					return nil, err
				}

				pkg.Files = strings.Split(string(content), "\n")
				pkg.Files = pkg.Files[1:]
				if pkg.Files[len(pkg.Files)-1] == "" {
					pkg.Files = pkg.Files[:len(pkg.Files)-1]
				}
			}
		}
	}

	if pkg != nil {
		pkgs = append(pkgs, pkg)
	}

	return pkgs, nil
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
					case `%DEPENDS%`:
						curr = pkg.depends
					case `%MAKEDEPENDS%`:
						curr = pkg.makedepends
					case `%OPTDEPENDS%`:
						curr = pkg.optdepends
					case ``:
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
