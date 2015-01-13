package main

import (
	"fmt"
	"github.com/CenturyLinkLabs/k8s-provision-vms/provision"
	"os"
	"strconv"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r)
			os.Exit(1)
		}
	}()

	c, _ := strconv.Atoi(os.Getenv("SERVER_COUNT"))

	params := provision.Params{Count: c}

	mIP := ""
	mPK := ""
	miIP := ""

	cp := provision.New("centurylink")
	s := cp.ProvisionCluster(params)
	for _, v := range s {
		if v.PrivateSSHKey == "" {
			miIP = miIP + "," + v.PrivateIP
		} else {
			mIP = v.PublicIP
			mPK = v.PrivateSSHKey
		}
	}

	setKey("REMOTE_TARGET_NAME", os.Getenv("REMOTE_TARGET_NAME"))
	setKey("API_KEY", os.Getenv("API_KEY"))
	setKey("API_PASSWORD", os.Getenv("API_PASSWORD"))
	setKey("OPEN_TCP_PORTS", os.Getenv("OPEN_TCP_PORTS"))
	setKey("KUBE_VERSION", os.Getenv("KUBE_VERSION"))
	setKey("REGION", os.Getenv("REGION"))

	setKey("MASTER_IP", mIP)
	setKey("MASTER_PRIVATE_KEY", mPK)
	setKey("MINION_IPS", miIP)
}
