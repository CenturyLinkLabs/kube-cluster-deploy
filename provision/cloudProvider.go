package provision

import "strings"

type CloudProvider interface {
	ProvisionCluster(params Params) []Server
}

func New(providerType string) CloudProvider {
	providerType = strings.ToLower(providerType)
	switch providerType {
	case "centurylink":
		return NewCenturyLink()
	}
	return nil
}

type Params struct {
	Count int
}
