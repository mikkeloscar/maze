package repo

import (
	"archive/tar"
	"bufio"
	"bytes"
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
	"sync"
	"time"

	"github.com/mikkeloscar/gopkgbuild"
	"github.com/mikkeloscar/maze/common/util"
	"github.com/mikkeloscar/maze/model"
)

var pkgPatt = regexp.MustCompile(`([a-z\d@._+]+[a-z\d@._+-]+)-((\d+:)?([\da-z\._+]+-\d+))-(i686|x86_64|any).pkg.tar.xz(.sig)?`)
var removePatt = regexp.MustCompile(`Removing existing entry '([a-z\d@._+-]+)'`)
var pkgNamePatt = regexp.MustCompile(`^[a-z\d@._+][a-z\d@._+-]*$`)

// ValidRepoName returns true if the name is a valid repo name.
// A valid name must only consist of lowercase alphanumerics and any of the
// following characters: @,.,_,+,-, and it must start with an alphanumeric.
func ValidRepoName(name string) bool {
	return pkgNamePatt.MatchString(name)
}

// ValidArchs returns true if all archs in the list are valid.
func ValidArchs(archs []string) bool {
	for _, arch := range archs {
		if !ValidArch(arch) {
			return false
		}
	}

	return true
}

// ValidArch returns true if the arch string is valid.
func ValidArch(arch string) bool {
	switch arch {
	case "x86_64":
		// fallthrough
		// case "i686":
		return true
	default:
		return false
	}
}

// Repo is a wrapper around the arch tools 'repo-add' and 'repo-remove'.
type Repo struct {
	*model.Repo
	basePath string
	Archs    []string // TODO: move to model.Repo
	rwLock   *sync.RWMutex
}

func NewRepo(r *model.Repo, basePath string) *Repo {
	return &Repo{r, basePath, []string{"x86_64"}, new(sync.RWMutex)}
}

func (r *Repo) InitDir() error {
	r.rwLock.Lock()
	defer r.rwLock.Unlock()

	for _, arch := range r.Archs {
		err := os.MkdirAll(r.PathDeep(arch), 0755)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Repo) ClearPath() error {
	r.rwLock.Lock()
	defer r.rwLock.Unlock()
	return os.RemoveAll(r.Path())
}

func (r *Repo) Path() string {
	return path.Join(r.basePath, r.Owner, r.Name)
}

func (r *Repo) PathDeep(arch string) string {
	return path.Join(r.Path(), arch)
}

// DB returns path to db archive.
func (r *Repo) DB(arch string) string {
	return path.Join(r.PathDeep(arch), fmt.Sprintf("%s.db.tar.gz", r.Name))
}

// FilesDB returns path to files db archive.
func (r *Repo) FilesDB(arch string) string {
	return path.Join(r.PathDeep(arch), fmt.Sprintf("%s.files.tar.gz", r.Name))
}

// InitEmptyDBs initialize empty dbs for the repo.
func (r *Repo) InitEmptyDBs() error {
	for _, arch := range r.Archs {
		args := []string{"--nocolor", "-R", r.DB(arch)}

		cmd := exec.Command("repo-add", args...)
		cmd.Dir = r.PathDeep(arch)

		err := cmd.Run()
		if err != nil {
			return err
		}
	}

	return nil
}

// Add adds a list of packages to a repo db, moving the package files to
// the repo db directory if needed.
func (r *Repo) Add(pkgPaths []string) error {
	if len(pkgPaths) == 0 {
		return nil
	}

	archPkgs := make(map[string][]string)

	for _, pkg := range pkgPaths {
		pkgPathDir, pkgPathBase := path.Split(pkg)
		_, _, arch, err := splitFileNameVersion(pkgPathBase)
		if err != nil {
			return err
		}

		archs := []string{arch}

		if arch == "any" {
			archs = r.Archs
		}

		for _, arch := range archs {
			if r.PathDeep(arch) != pkgPathDir {
				// move pkg to repo path.
				newPath := path.Join(r.PathDeep(arch), pkgPathBase)
				err := os.Rename(pkg, newPath)
				if err != nil {
					return err
				}
				archPkgs[arch] = append(archPkgs[arch], newPath)
			}
		}
	}

	r.rwLock.Lock()
	defer r.rwLock.Unlock()
	for arch, pkgs := range archPkgs {
		args := []string{"--nocolor", "-R", r.DB(arch)}
		args = append(args, pkgs...)

		var stderr bytes.Buffer
		cmd := exec.Command("repo-add", args...)
		cmd.Dir = r.PathDeep(arch)
		cmd.Stderr = &stderr

		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("%s: %s", err, stderr.String())
		}
	}

	return nil
}

// Remove removes a list of packages from the repo db.
func (r *Repo) Remove(pkgs []string, arch string) error {
	args := []string{"--nocolor", r.DB(arch)}
	args = append(args, pkgs...)

	cmd := exec.Command("repo-remove", args...)
	cmd.Dir = r.PathDeep(arch)

	r.rwLock.Lock()
	defer r.rwLock.Unlock()
	out, err := cmd.Output()
	if err != nil {
		return err
	}

	// remove repo file
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if !strings.HasPrefix(line, "  ->") {
			continue
		}

		match := removePatt.FindStringSubmatch(line)
		if len(match) == 2 {
			fs, err := ioutil.ReadDir(r.PathDeep(arch))
			if err != nil {
				return err
			}

			for _, f := range fs {
				if !strings.HasPrefix(f.Name(), match[1]) {
					continue
				}

				err := os.Remove(path.Join(r.PathDeep(arch), f.Name()))
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// IsNewFilename returns true if pkgfile is a newer version than what's in the
// repo.
// If the package is not found in the repo, it will be marked as new.
func (r *Repo) IsNewFilename(file string) (bool, error) {
	name, arch, version, err := splitFileNameVersion(file)
	if err != nil {
		return false, err
	}

	ver, err := pkgbuild.NewCompleteVersion(version)
	if err != nil {
		return false, err
	}

	return r.IsNew(name, arch, *ver)
}

// IsNew returns true if pkg is a newer version than what's in the repo.
// If the package is not found in the repo, it will be marked as new.
func (r *Repo) IsNew(name, arch string, version pkgbuild.CompleteVersion) (bool, error) {
	r.rwLock.RLock()
	defer r.rwLock.RUnlock()

	archs := []string{arch}

	if arch == "any" {
		archs = r.Archs
	}

Loop:
	for _, arch := range archs {

		f, err := os.Open(r.DB(arch))
		if err != nil {
			if os.IsNotExist(err) {
				break
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
						break Loop
					}
					return false, nil
				}
			case tar.TypeReg:
				continue
			}
		}
	}

	return true, nil
}

// Obsolete returns a list of obsolete packages based on the input packages.
// A package is considered obsolete if it's not in the input list and not a
// dependency of one of the input packages.
func (r *Repo) Obsolete(pkgs []string, arch string) ([]string, error) {
	pkgMap, err := r.readPkgMap(arch)
	if err != nil {
		return nil, err
	}

	return r.obsolete(pkgs, pkgMap), nil
}

func (r *Repo) obsolete(pkgs []string, pkgMap map[string]*pkgDep) []string {
	var obsolete map[string]struct{}

	for n := range pkgMap {
		if !util.StrContains(n, pkgs) {
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
func (r *Repo) Package(name, arch string, files bool) (*model.Package, error) {
	r.rwLock.RLock()
	defer r.rwLock.RUnlock()

	f, err := os.Open(r.FilesDB(arch))
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
	pkg := &model.Package{
		Depends:     []string{},
		OptDepends:  []string{},
		MakeDepends: []string{},
		Files:       []string{},
	}

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

	if pkg.Name == "" {
		return nil, nil
	}

	return pkg, nil
}

// Packages returns a list of all packages in the repo.
func (r *Repo) Packages(arch string, files bool) ([]*model.Package, error) {
	r.rwLock.RLock()
	defer r.rwLock.RUnlock()

	var pkgs []*model.Package

	f, err := os.Open(r.FilesDB(arch))
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

			pkg = &model.Package{
				Depends:     []string{},
				OptDepends:  []string{},
				MakeDepends: []string{},
				Files:       []string{},
			}
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

func (r *Repo) readPkgMap(arch string) (map[string]*pkgDep, error) {
	pkgMap := make(map[string]*pkgDep)

	f, err := os.Open(r.DB(arch))
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

// turn "zlib-1.2.8-4-x86_64.pkg.tar.xz" into ("zlib", "1.2.8-4", "x86_64").
func splitFileNameVersion(file string) (string, string, string, error) {
	match := pkgPatt.FindStringSubmatch(file)
	if len(match) > 0 {
		return match[1], match[2], match[5], nil
	}

	return "", "", "", fmt.Errorf("invalid package filename")
}
