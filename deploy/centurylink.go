package deploy

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/CenturyLinkLabs/clcgo"
	"github.com/CenturyLinkLabs/kube-cluster-deploy/utils"
	"golang.org/x/crypto/ssh"
	"net"
	"time"
)

type CenturyLink struct {
	clcClient      *clcgo.Client
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
		return errors.New("\nMissing values to create cluster. Check documentation for required values.")
	}

	clc.clcClient = clcgo.NewClient()
	if clc.ServerTemplate == "" {
		clc.ServerTemplate = "RHEL-7-64-TEMPLATE"
	}

	e := clc.clcClient.GetAPICredentials(clc.APIUsername, clc.APIPassword)
	if e != nil {
		return e
	}

	return nil
}

func (clc *CenturyLink) createServer() (CloudServer, error) {

	utils.LogInfo("\nDeploying Server")

	s := clcgo.Server{
		Name:           clc.ServerName,
		GroupID:        clc.GroupID,
		SourceServerID: clc.ServerTemplate,
		CPU:            clc.CPU,
		MemoryGB:       clc.MemoryGB,
		Type:           "standard",
	}

	st, e := clc.clcClient.SaveEntity(&s)
	if e != nil {
		return CloudServer{}, e
	}

	utils.LogInfo("\nWaiting for server to provision...")
	e = clc.waitForJob(st)
	if e != nil {
		return CloudServer{}, e
	}
	clc.clcClient.GetEntity(&s)

	e = clc.addPublicIP(s)
	if e != nil {
		return CloudServer{}, e
	}
	clc.clcClient.GetEntity(&s)

	utils.LogInfo("\nServer is provisioned: " + s.Name)

	cr := clcgo.Credentials{Server: s}
	clc.clcClient.GetEntity(&cr)

	pubIP := clc.publicIPFromServer(s)
	priIP := clc.privateIPFromServer(s)

	priKey := clc.PrivateSSHKey
	utils.LogInfo(fmt.Sprintf("\nPublicIP: %s, PrivateIP: %s", pubIP, priIP))

	clc.addSSHKey(pubIP, cr.Password, clc.PublicSSHKey, priKey)

	pmxS := CloudServer{
		Name:          s.Name,
		PublicIP:      pubIP,
		PrivateIP:     priIP,
		PublicSSHKey:  clc.PublicSSHKey,
		PrivateSSHKey: priKey,
	}

	utils.LogInfo("Server deployment complete!!")

	return pmxS, nil
}

func (clc *CenturyLink) addPublicIP(s clcgo.Server) error {

	var ps []clcgo.Port
	for _, p := range clc.TCPOpenPorts {
		ps = append(ps, clcgo.Port{Protocol: "TCP", Port: p})
	}
	ps = append(ps, clcgo.Port{Protocol: "TCP", Port: 22})

	priIP := clc.privateIPFromServer(s)

	a := clcgo.PublicIPAddress{Server: s, Ports: ps, InternalIPAddress: priIP}
	st, e := clc.clcClient.SaveEntity(&a)
	if e != nil {
		return e
	}

	utils.LogInfo("Adding public IP...")
	e = clc.waitForJob(st)
	if e != nil {
		return e
	}

	utils.LogInfo("Public IP is added!")
	return nil
}

func (clc *CenturyLink) addSSHKey(publicIp string, password string, pubKey string, privateKey string) {

	utils.LogInfo("\nWaiting for server to start before adding ssh keys...")
	clc.WaitForTCP(publicIp)

	utils.LogInfo("\nServer Up....Adding SSH keys")
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
	utils.LogInfo("\nSSH Keys added!")
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
	utils.LogInfo("\nWaiting for server to start...")
	for {
		conn, err := net.Dial("tcp", addr+":22")
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

func (clc *CenturyLink) waitForJob(st clcgo.Status) error {
	for !st.HasSucceeded() {
		time.Sleep(time.Second * 10)
		e := clc.clcClient.GetEntity(&st)
		if e != nil {
			return e
		}
	}
	return nil
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
