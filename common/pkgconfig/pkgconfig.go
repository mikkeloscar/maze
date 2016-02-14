package pkgconfig

import (
	"io"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// PkgConfig defines the packages to be build for a repository.
type PkgConfig struct {
	AUR []string `yaml:"aur"`
}

// ReadConfig reads the content of an io.ReadCloser into a PkgConfig struct.
func ReadConfig(content io.ReadCloser) (*PkgConfig, error) {
	data, err := ioutil.ReadAll(content)
	if err != nil {
		return nil, err
	}

	config := new(PkgConfig)

	err = yaml.Unmarshal(data, config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
