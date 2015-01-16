package provision

import (
	"fmt"
	"github.com/CenturyLinkLabs/clcgo"
	"github.com/CenturyLinkLabs/k8s-provision-vms/deploy"
	"github.com/CenturyLinkLabs/k8s-provision-vms/utils"
	"os"
	"strconv"
)

type CenturyLink struct {
	clcClient           *clcgo.Client
	networkName         string
	groupId             string
	cpu                 int
	memGb               int
	masterPrivateSSHKey string
	masterPublicSSHKey  string
	uname               string
	password            string
}

func NewCenturyLink() *CenturyLink {
	cl := new(CenturyLink)
	return cl
}

func (clc CenturyLink) ProvisionCluster(params Params) ([]Server, error) {
	fmt.Printf("\nProvisioning Server Cluster in Centurylink")
	clc.initProvider()

	fmt.Printf("\nMinion Count: %d", params.MinionCount)
	var servers []Server
	for i := 0; i < params.MinionCount+1; i++ {

		pk := ""
		if i == 0 {
			fmt.Printf("\nDeploying Kubernetes Master")
			pk = clc.masterPrivateSSHKey
		} else {
			fmt.Printf("\nDeploying Kubernetes Minion %d", i)
		}

		c := deploy.CenturyLink{
			PrivateSSHKey: pk,
			PublicSSHKey:  clc.masterPublicSSHKey,
			APIUsername:   clc.uname,
			APIPassword:   clc.password,
			GroupID:       clc.groupId,
			CPU:           clc.cpu,
			MemoryGB:      clc.memGb,
			ServerName:    "KUBE",
		}

		s, e := c.DeployVM()
		if e != nil {
			panic(e)
			return nil, e
		}

		servers = append(servers, Server{Name: s.Name, PublicIP: s.PublicIP, PrivateIP: s.PrivateIP, PublicSSHKey: s.PublicSSHKey, PrivateSSHKey: pk})
	}
	return servers, nil
}

func (clc *CenturyLink) initProvider() bool {
	clc.uname = os.Getenv("USERNAME")
	clc.password = os.Getenv("PASSWORD")
	clc.networkName = os.Getenv("NETWORK_NAME")
	clc.groupId = os.Getenv("GROUP_ID")
	clc.cpu, _ = strconv.Atoi(os.Getenv("CPU"))
	clc.memGb, _ = strconv.Atoi(os.Getenv("MEMORY_GB"))

	if clc.uname == "" || clc.password == "" || clc.networkName == "" {
		fmt.Printf("\n\nMissing Params.. in cluster creation...Check Docs....\n\n")
	}

	clc.masterPrivateSSHKey, clc.masterPublicSSHKey, _ = utils.CreateSSHKey()

	return true
}
