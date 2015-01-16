package deploy

import (
	"bytes"
	"fmt"
	"github.com/CenturyLinkLabs/clcgo"
	"golang.org/x/crypto/ssh"
	"net"
	"time"
)

type CenturyLink struct {
	clcClient      *clcgo.Client
	NetworkName    string
	CPU            int
	MemoryGB       int
	PrivateSSHKey  string
	PublicSSHKey   string
	APIUsername    string
	APIPassword    string
	TCPOpenPorts   []int
	GroupID        string
	ServerName     string
	ServerTemplate string
}

func NewCenturyLink() *CenturyLink {
	cl := new(CenturyLink)
	return cl
}

func (clc CenturyLink) DeployVM() (CloudServer, error) {
	e := clc.initProvider()
	if e != nil {
		return CloudServer{}, e
	}
	return clc.createServer()
}

func (clc *CenturyLink) initProvider() error {

	if clc.APIUsername == "" || clc.APIPassword == "" || clc.GroupID == "" {
		fmt.Print("\n\nMissing Params...Check Docs....\n\n")
	}

	clc.clcClient = clcgo.NewClient()
	clc.ServerTemplate = "RHEL-7-64-TEMPLATE"

	e := clc.clcClient.GetAPICredentials(clc.APIUsername, clc.APIPassword)

	if e != nil {
		return e
	}
	return nil
}

func (clc *CenturyLink) createServer() (CloudServer, error) {
	s := clcgo.Server{
		Name:           clc.ServerName,
		GroupID:        clc.GroupID,
		SourceServerID: clc.ServerTemplate,
		CPU:            clc.CPU,
		MemoryGB:       clc.MemoryGB,
		Type:           "standard",
	}

	fmt.Print("\nDeploying Server")
	st, e := clc.clcClient.SaveEntity(&s)
	if e != nil {
		panic(e)
		return CloudServer{}, e
	}

	fmt.Print("\nWaiting for server to provision")
	for !st.HasSucceeded() {
		time.Sleep(time.Second * 10)
		clc.clcClient.GetEntity(st)
	}

	clc.clcClient.GetEntity(&s)

	fmt.Printf("\nServer is provisioned: %s", s.Name)

	var ports []clcgo.Port
	for p, _ := range clc.TCPOpenPorts {
		ports = append(ports, clcgo.Port{Protocol: "TCP", Port: p})
	}
	ports = append(ports, clcgo.Port{Protocol: "TCP", Port: 22})
	ports = append(ports, clcgo.Port{Protocol: "TCP", Port: 8080})

	a := clcgo.PublicIPAddress{Server: s, Ports: ports}
	st, e = clc.clcClient.SaveEntity(&a)
	if e != nil {
		return CloudServer{}, e
	}

	fmt.Print("Adding public IP")
	for !st.HasSucceeded() {
		time.Sleep(time.Second * 10)
		clc.clcClient.GetEntity(st)
	}

	fmt.Print("\nPublic IP is added!")
	clc.clcClient.GetEntity(&s)

	cr := clcgo.Credentials{Server: s}
	clc.clcClient.GetEntity(&cr)

	pubIp := clc.publicIPFromServer(s)
	priIp := clc.privateIPFromServer(s)

	fmt.Printf("PublicIP: %s, PrivateIp: %s", pubIp, priIp)

	clc.WaitForTCP(pubIp)

	fmt.Print("Server Up....Adding SSH keys")

	priKey := clc.PrivateSSHKey

	clc.addSSHKey(pubIp, cr.Password, clc.PublicSSHKey, priKey)
	pmxS := CloudServer{
		Name:          s.Name,
		PublicIP:      pubIp,
		PrivateIP:     priIp,
		PublicSSHKey:  clc.PublicSSHKey,
		PrivateSSHKey: priKey,
	}
	println("\nServer deployment complete")

	return pmxS, nil
}

func (clc *CenturyLink) addSSHKey(publicIp string, password string, pubKey string, privateKey string) {

	config := &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{ssh.Password(password)},
	}

	cmd := fmt.Sprintf("echo -e \"%s\" >> ~/.ssh/authorized_keys", pubKey)
	clc.executeCmd(cmd, publicIp, config)

	if privateKey != "" {
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
