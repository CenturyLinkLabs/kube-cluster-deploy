package provision

import (
	"bytes"
	"fmt"
	"github.com/CenturyLinkLabs/clcgo"
	"golang.org/x/crypto/ssh"
	"net"
	"os"
	"strconv"
	"time"
)

type CenturyLink struct {
	clcClient           *clcgo.Client
	networkName         string
	groupId             string
	cpu                 int
	memGb               int
	masterPrivateSSHKey string
	masterPublicSSHKey  string
}

func NewCenturyLink() *CenturyLink {
	cl := new(CenturyLink)
	return cl
}

func (clc CenturyLink) ProvisionCluster(params Params) []Server {
	println("\nProvisioning Server Cluster in Centurylink")
	clc.initProvider()

	println("Deploying ", params.MinionCount, " Servers")
	var servers []Server
	for i := 0; i < params.MinionCount+1; i++ {
		println("Deploying server ", i+1)
		s := clc.createServer(i + 1)
		servers = append(servers, s)
	}
	return servers
}

func (clc *CenturyLink) initProvider() bool {
	uname := os.Getenv("USERNAME")
	password := os.Getenv("PASSWORD")
	clc.networkName = os.Getenv("NETWORK_NAME")
	clc.groupId = os.Getenv("GROUP_ID")
	clc.cpu, _ = strconv.Atoi(os.Getenv("CPU"))
	clc.memGb, _ = strconv.Atoi(os.Getenv("MEMORY_GB"))

	if uname == "" || password == "" || clc.networkName == "" {
		println("\n\nMissing Params...Check Docs....\n\n")
	}

	clc.clcClient = clcgo.NewClient()
	err := clc.clcClient.GetAPICredentials(uname, password)
	println("Logged in")

	if err != nil {
		panic(err)
	}

	clc.masterPrivateSSHKey, clc.masterPublicSSHKey = createSSHKey()

	return true
}

func (clc *CenturyLink) createServer(serverNumber int) Server {
	s := clcgo.Server{
		Name:           "KUBE",
		GroupID:        clc.groupId,
		SourceServerID: "RHEL-7-64-TEMPLATE",
		CPU:            clc.cpu,
		MemoryGB:       clc.memGb,
		//NetworkID:      "ad88c21819c84b788dc84f27989ade40",
		Type: "standard",
	}

	st, err := clc.clcClient.SaveEntity(&s)
	if err != nil {
		println(err.Error())
	}

	println("Waiting for server to provision")
	for !st.HasSucceeded() {
		time.Sleep(time.Second * 10)
		print(" .")
		clc.clcClient.GetEntity(st)
	}

	clc.clcClient.GetEntity(&s)

	println("\nServer is provisioned:", s.Name)

	ports := []clcgo.Port{clcgo.Port{Protocol: "TCP", Port: 22}}
	a := clcgo.PublicIPAddress{Server: s, Ports: ports}
	st, err = clc.clcClient.SaveEntity(&a)
	if err != nil {
		println(err.Error())
	}

	println("Adding public IP")
	for !st.HasSucceeded() {
		time.Sleep(time.Second * 10)
		print(" .")
		clc.clcClient.GetEntity(st)
	}

	println("\nPublic IP is added!")
	clc.clcClient.GetEntity(&s)

	cr := clcgo.Credentials{Server: s}
	clc.clcClient.GetEntity(&cr)

	pubIp := clc.publicIPFromServer(s)
	priIp := clc.privateIPFromServer(s)

	clc.WaitForTCP(pubIp)

	println("Server Up....Adding SSH keys")

	master := false
	if serverNumber == 1 {
		master = true
	}

	priKey := clc.masterPrivateSSHKey
	if !master {
		priKey = ""
	}

	clc.addSSHKey(pubIp, cr.Password, clc.masterPublicSSHKey, priKey, master)
	pmxS := Server{
		Name:          s.Name,
		PublicIP:      pubIp,
		PrivateIP:     priIp,
		PublicSSHKey:  clc.masterPublicSSHKey,
		PrivateSSHKey: priKey,
	}
	println("\nServer deployment complete")

	return pmxS
}

func (clc *CenturyLink) addSSHKey(publicIp string, password string, pubKey string, privateKey string, master bool) {

	config := &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{ssh.Password(password)},
	}

	cmd := fmt.Sprintf("echo -e \"%s\" >> ~/.ssh/authorized_keys", pubKey)
	clc.executeCmd(cmd, publicIp, config)

	if master {
		pKCmd := fmt.Sprintf("echo -e \"%s\" >> ~/.ssh/id_rsa && chmod 400 ~/.ssh/id_rsa", privateKey)
		clc.executeCmd(pKCmd, publicIp, config)
	}
}

func (clc *CenturyLink) publicIPFromServer(s clcgo.Server) string {
	addresses := s.Details.IPAddresses
	for _, a := range addresses {
		if a.Public != "" {
			return a.Public
		}
	}
	return ""
}

func (clc *CenturyLink) privateIPFromServer(s clcgo.Server) string {
	addresses := s.Details.IPAddresses
	for _, a := range addresses {
		if a.Internal != "" {
			return a.Internal
		}
	}
	return ""
}

func (clc *CenturyLink) executeCmd(cmd, hostname string, config *ssh.ClientConfig) string {
	conn, _ := ssh.Dial("tcp", hostname+":22", config)
	session, _ := conn.NewSession()
	defer session.Close()

	var stdoutBuf bytes.Buffer
	session.Stdout = &stdoutBuf
	session.Run(cmd)

	return hostname + ": " + stdoutBuf.String()
}

func (clc *CenturyLink) WaitForTCP(addr string) error {
	println("Waiting for server to start")
	for {
		conn, err := net.Dial("tcp", addr+":22")
		print(" . ")
		if err != nil {
			continue
		}
		defer conn.Close()
		if _, err = conn.Read(make([]byte, 1)); err != nil {
			continue
		}
		break
	}
	return nil
}
