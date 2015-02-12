package provision

import "strings"
import "github.com/CenturylinkLabs/kube-cluster-deploy/deploy"

// CloudProvider is used to deploy kubernetes cluster on any of the supported
// cloud providers.
type CloudProvider interface {
	ProvisionCluster() ([]deploy.CloudServer, error)
}

// New is used to instantiate a CloudProvider to use to provision the cluster.
func New(providerType string) CloudProvider {
	providerType = strings.ToLower(providerType)
	switch providerType {
	case "centurylink":
		return NewCenturylink()
    case "amazon":
        return NewAmazon()
	}
	return nil
}

// Params are the common params that are passed to all the cloud provider.
// Specific params are passed as environment variables.
type Params struct {
	MinionCount int
}
