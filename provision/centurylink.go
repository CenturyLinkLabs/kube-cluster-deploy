package provision

import (
	"errors"
	"github.com/CenturyLinkLabs/clcgo"
	"github.com/CenturyLinkLabs/kube-cluster-deploy/deploy"
	"github.com/CenturyLinkLabs/kube-cluster-deploy/utils"
	"os"
	"strconv"
	"strings"
)

// CenturyLink has the data that is used for provisioning a server. Most of the
// data is passed in environment variables. The following env vars are required
// for provisioning a server in CenturyLink, USERNAME, PASSWORD, GROUP_ID, CPU,
// MEMORY_GB, OPEN_TCP_PORTS
type CenturyLink struct {
	clcClient   *clcgo.Client
	groupID     string
	cpu         int
	memGb       int
	masterPK    string
	masterPuK   string
	uname       string
	password    string
	minionPorts []int
}

// NewCenturyLink is used to create a new client for using CenturyLink struct to
// create servers.
func NewCenturyLink() *CenturyLink {
	cl := new(CenturyLink)
	return cl
}

// ProvisionCluster is used to provision a cluster of RHEL7 VMs (1 Master +
// n Minions).
func (clc CenturyLink) ProvisionCluster(params Params) ([]Server, error) {
	utils.LogInfo("\nProvisioning Server Cluster into Centurylink Cloud...")
	utils.LogInfo("\nMinion Count: " + strconv.Itoa(params.MinionCount))

	e := clc.initProvider()
	if e != nil {
		return nil, e
	}

	var servers []Server
	for i := 0; i < params.MinionCount+1; i++ {
		pk := ""
		if i == 0 {
			utils.LogInfo("\nDeploying Kubernetes Master...")
			pk = clc.masterPK
		} else {
			utils.LogInfo("\nDeploying Kubernetes Minion... " + strconv.Itoa(i))
		}

		c := deploy.CenturyLink{
			PrivateSSHKey: pk,
			PublicSSHKey:  clc.masterPuK,
			APIUsername:   clc.uname,
			APIPassword:   clc.password,
			GroupID:       clc.groupID,
			CPU:           clc.cpu,
			MemoryGB:      clc.memGb,
			ServerName:    "KUBE",
			TCPOpenPorts:  clc.minionPorts,
		}

		s, e := c.DeployVM()
		if e != nil {
			return nil, e
		}

		servers = append(servers, Server{Name: s.Name, PublicIP: s.PublicIP, PrivateIP: s.PrivateIP, PublicSSHKey: s.PublicSSHKey, PrivateSSHKey: pk})
	}
	return servers, nil
}

func (clc *CenturyLink) initProvider() error {
	clc.uname = os.Getenv("USERNAME")
	clc.password = os.Getenv("PASSWORD")
	clc.groupID = os.Getenv("GROUP_ID")
	clc.cpu, _ = strconv.Atoi(os.Getenv("CPU"))
	clc.memGb, _ = strconv.Atoi(os.Getenv("MEMORY_GB"))
	ps := os.Getenv("OPEN_TCP_PORTS")
	if ps != "" {
		s := strings.Split(ps, ",")
		for _, p := range s {
			v, e := strconv.Atoi(p)
			if e == nil {
				clc.minionPorts = append(clc.minionPorts, v)
			}
		}
	}

	if clc.uname == "" || clc.password == "" || clc.groupID == "" {
		return errors.New("\n\nMissing values to create cluster. Check documentation for required values\n\n")
	}

	if clc.cpu <= 0 || clc.memGb <= 0 {
		return errors.New("\n\nMake sure CPU & MemoryGB values are greater than 0.\n\n")
	}

	pk, puk, err := utils.CreateSSHKey()
	clc.masterPK = pk
	clc.masterPuK = puk

	if err != nil {
		return err
	}

	return nil

}
