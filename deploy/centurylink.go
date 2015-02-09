package deploy

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/CenturylinkLabs/clcgo"
	"github.com/CenturylinkLabs/kube-cluster-deploy/utils"
	"golang.org/x/crypto/ssh"
	"time"
)

type Centurylink struct {
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

func NewCenturylink() *Centurylink {
	cl := new(Centurylink)
	return cl
}

func (clc Centurylink) DeployVMs() ([]CloudServer, error) {
	e := clc.initProvider()
	if e != nil {
		return nil, e
	}

    s, e := clc.createServer()
    var servers []CloudServer
    servers = append(servers, s )
    return servers, e
}

func (clc *Centurylink) initProvider() error {

	if clc.APIUsername == "" || clc.APIPassword == "" || clc.GroupID == "" {
		return errors.New("\nMissing Params...Check Docs....")
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

func (clc *Centurylink) createServer() (CloudServer, error) {

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

	utils.LogInfo("\nWaiting for server to provision")
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

func (clc *Centurylink) addPublicIP(s clcgo.Server) error {

	var ps []clcgo.Port
    clc.TCPOpenPorts = append(clc.TCPOpenPorts, 22)
	for _, p := range clc.TCPOpenPorts {
		ps = append(ps, clcgo.Port{Protocol: "TCP", Port: p})
	}
	priIP := clc.privateIPFromServer(s)

	a := clcgo.PublicIPAddress{Server: s, Ports: ps, InternalIPAddress: priIP}
	st, e := clc.clcClient.SaveEntity(&a)
	if e != nil {
		return e
	}

	utils.LogInfo("Adding public IP")
	e = clc.waitForJob(st)
	if e != nil {
		return e
	}

	utils.LogInfo("Public IP is added!")
	return nil
}

func (clc *Centurylink) addSSHKey(publicIp string, password string, pubKey string, privateKey string) {

	utils.LogInfo("\nWaiting for server to start before adding ssh keys")
	utils.WaitForSSH(publicIp)

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
	utils.LogInfo("\nSSH Keys added")
}

func (clc *Centurylink) executeCmd(cmd, hostname string, config *ssh.ClientConfig) string {
	conn, _ := ssh.Dial("tcp", hostname+":22", config)
	session, _ := conn.NewSession()
	defer session.Close()

	var stdoutBuf bytes.Buffer
	session.Stdout = &stdoutBuf
	session.Run(cmd)

	return hostname + ": " + stdoutBuf.String()
}


func (clc *Centurylink) waitForJob(st clcgo.Status) error {
	for !st.HasSucceeded() {
		time.Sleep(time.Second * 10)
		e := clc.clcClient.GetEntity(&st)
		if e != nil {
			return e
		}
	}
	return nil
}

func (clc *Centurylink) publicIPFromServer(s clcgo.Server) string {
	addresses := s.Details.IPAddresses
	for _, a := range addresses {
		if a.Public != "" {
			return a.Public
		}
	}
	return ""
}

func (clc *Centurylink) privateIPFromServer(s clcgo.Server) string {
	addresses := s.Details.IPAddresses
	for _, a := range addresses {
		if a.Internal != "" {
			return a.Internal
		}
	}
	return ""
}
