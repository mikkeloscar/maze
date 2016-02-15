package model

import "time"

type Package struct {
	FileName    string    `json:"filename"`
	Name        string    `json:"name"`
	Base        string    `json:"base"`
	Version     string    `json:"version"`
	Desc        string    `json:"desc"`
	CSize       string    `json:"csize"`
	ISize       string    `json:"isize"`
	MD5Sum      string    `json:"md5sum"`
	SHA256Sum   string    `json:"sha256sum"`
	URL         string    `json:"url"`
	License     string    `json:"license"`
	Arch        string    `json:"arch"`
	BuildDate   time.Time `json:"build_date"`
	Packager    string    `json:"packager"`
	Depends     []string  `json:"depends"`
	OptDepends  []string  `json:"optdepends"`
	MakeDepends []string  `json:"makedpends"`
	Files       []string  `json:"files"`
}
