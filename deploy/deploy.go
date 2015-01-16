package deploy

import (
	"errors"
	"strings"
)

type Deploy interface {
	DeployVM() (CloudServer, error)
}

func New(p string) (Deploy, error) {
	switch strings.ToLower(p) {
	case "centurylink":
		return NewCenturyLink(), nil
	}
	return nil, errors.New("Unsupported Cloud Provider")
}

type CloudServer struct {
	Name          string
	PublicIP      string
	PrivateIP     string
	PublicSSHKey  string
	PrivateSSHKey string
}
