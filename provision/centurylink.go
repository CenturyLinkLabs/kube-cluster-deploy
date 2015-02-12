package provision

import (
	"errors"
	"github.com/CenturylinkLabs/clcgo"
	"github.com/CenturylinkLabs/kube-cluster-deploy/deploy"
	"github.com/CenturylinkLabs/kube-cluster-deploy/utils"
	"os"
	"strconv"
	"strings"
    )

// Centurylink has the data that is used for provisioning a server. Most of the
// data is passed in environment variables. The following env vars are required
// for provisioning a server in Centurylink, USERNAME, PASSWORD, GROUP_ID, CPU,
// MEMORY_GB, OPEN_TCP_PORTS
type Centurylink struct {
	clcClient   *clcgo.Client
	gID     string
	cpu         int
	memGb       int
	masterPK    string
	masterPuK   string
	uname       string
	password    string
	minionPorts []int
}

// NewCenturylink is used to create a new client for using Centurylink client to
// create RHEL 7 server cluster.
func NewCenturylink() *Centurylink {
	cl := new(Centurylink)
	return cl
}

// ProvisionCluster is used to provision a cluster of RHEL7 VMs (1 Master +
// n Minions).
func (clc Centurylink) ProvisionCluster() ([]deploy.CloudServer, error) {
	utils.LogInfo("\nProvisioning Server Cluster in Centurylink")

    cnt, e := strconv.Atoi(os.Getenv("MINION_COUNT"))
	utils.LogInfo("\nMinion Count: " + strconv.Itoa(cnt))

	e = clc.initProvider()
	if e != nil {
		return nil, e
	}

	var servers []deploy.CloudServer
	for i := 0; i < cnt+1; i++ {
		pk := ""
		if i == 0 {
			utils.LogInfo("\nDeploying Kubernetes Master...")
			pk = clc.masterPK
		} else {
			utils.LogInfo("\nDeploying Kubernetes Minion... " + strconv.Itoa(i))
		}

		c := deploy.Centurylink{
			PrivateSSHKey: pk,
			PublicSSHKey:  clc.masterPuK,
			APIUsername:   clc.uname,
			APIPassword:   clc.password,
			GroupID:       clc.gID,
			CPU:           clc.cpu,
			MemoryGB:      clc.memGb,
			ServerName:    "KUBE",
			TCPOpenPorts:  clc.minionPorts,
		}

		s, e := c.DeployVMs()
		if e != nil {
			return nil, e
		}

        s[0].PrivateSSHKey = pk
        s[0].PublicSSHKey = clc.masterPuK

		servers = append(servers, s[0])
	}
	return servers, nil
}

func (clc *Centurylink) initProvider() error {
	clc.uname = os.Getenv("USERNAME")
	clc.password = os.Getenv("PASSWORD")
	clc.gID = os.Getenv("GROUP_ID")
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

	if clc.uname == "" || clc.password == "" || clc.gID == "" {
		return errors.New("\n\nMissing Params.. in cluster creation...Check Docs....\n\n")
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
