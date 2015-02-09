package deploy

import (
	"errors"
	"strings"
)

type Deploy interface {
	DeployVMs() ([]CloudServer, error)
}

func New(p string) (Deploy, error) {
	switch strings.ToLower(p) {
	case "centurylink":
		return NewCenturylink(), nil
	}
	return nil, errors.New("Unsupported Cloud Provider")
}

type Params struct {
    Count int
}

type CloudServer struct {
	Name          string
	PublicIP      string
	PrivateIP     string
	PublicSSHKey  string
	PrivateSSHKey string
}
