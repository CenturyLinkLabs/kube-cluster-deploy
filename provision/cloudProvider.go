package provision

import "strings"

// CloudProvider is used to deploy kubernetes cluster on any of the supported
// cloud providers.
type CloudProvider interface {
	ProvisionCluster(params Params) ([]Server, error)
}

// New is used to instantiate a CloudProvider to use to provision the cluster.
func New(providerType string) CloudProvider {
	providerType = strings.ToLower(providerType)
	switch providerType {
	case "centurylink":
		return NewCenturyLink()
	}
	return nil
}

// Params are the common params that are passed to all the cloud provider.
// Specific params are passed as environment variables.
type Params struct {
	MinionCount int
}

// A Server array is returned once the cluster is provisioned and has necessary
// information to connect to the server.
type Server struct {
	Name          string
	PublicIP      string
	PrivateIP     string
	PublicSSHKey  string
	PrivateSSHKey string
}
