package main

import (
	"encoding/base64"
	"fmt"
	"github.com/CenturyLinkLabs/k8s-provision-vms/provision"
	"github.com/CenturyLinkLabs/k8s-provision-vms/utils"
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
	s, e := cp.ProvisionCluster(params)

	if e != nil {
		panic(e)
	}

	for _, v := range s {
		if v.PrivateSSHKey == "" {
			miIP = append(miIP, v.PrivateIP)
		} else {
			mIP = v.PublicIP
			mPK = v.PrivateSSHKey
		}
	}

	utils.SetKey("MASTER_IP", mIP)
	utils.SetKey("MASTER_PRIVATE_KEY", base64.StdEncoding.EncodeToString([]byte(mPK)))
	utils.SetKey("MINION_IPS", strings.Join(miIP, ","))
}
