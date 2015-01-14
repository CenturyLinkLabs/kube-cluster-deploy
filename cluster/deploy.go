package main

import (
	"encoding/base64"
	"fmt"
	"github.com/CenturyLinkLabs/k8s-provision-vms/provision"
	"os"
	"strconv"
	"strings"
)

func main() {

	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r)
			os.Exit(1)
		}
	}()

	mIP := ""
	mPK := ""
	var miIP []string

	c, _ := strconv.Atoi(os.Getenv("MINION_COUNT"))
	params := provision.Params{MinionCount: c}
	cp := provision.New("centurylink")
	s := cp.ProvisionCluster(params)
	for _, v := range s {
		if v.PrivateSSHKey == "" {
			miIP = append(miIP, v.PrivateIP)
		} else {
			mIP = v.PublicIP
			mPK = v.PrivateSSHKey
		}
	}

	setKey("REMOTE_TARGET_NAME", os.Getenv("REMOTE_TARGET_NAME"))
	setKey("API_KEY", os.Getenv("USERNAME"))
	setKey("API_PASSWORD", os.Getenv("PASSWORD"))
	setKey("OPEN_TCP_PORTS", os.Getenv("OPEN_TCP_PORTS"))
	setKey("KUBE_VERSION", os.Getenv("KUBE_VERSION"))
	setKey("REGION", os.Getenv("REGION"))
	setKey("NETWORK_NAME", os.Getenv("NETWORK_NAME"))
	setKey("MASTER_IP", mIP)
	setKey("MASTER_PRIVATE_KEY", base64.StdEncoding.EncodeToString([]byte(mPK)))
	setKey("MINION_IPS", strings.Join(miIP, ","))
}
