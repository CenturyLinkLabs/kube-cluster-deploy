package main

import (
	"encoding/base64"
	"fmt"
	"github.com/CenturyLinkLabs/kube-cluster-deploy/provision"
	"github.com/CenturyLinkLabs/kube-cluster-deploy/utils"
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
	mPrIP := ""
	mPK := ""
	var miIP []string

	c, _ := strconv.Atoi(os.Getenv("MINION_COUNT"))

	if c == 0 {
		panic("\nPlease make sure you have at least one minion in the cluster.")
	}

	params := provision.Params{MinionCount: c}

	cp := provision.New("centurylink")
	s, e := cp.ProvisionCluster(params)

	if e != nil {
		panic(e.Error())
	}

	for _, v := range s {
		if v.PrivateSSHKey == "" {
			miIP = append(miIP, v.PrivateIP)
		} else {
			mIP = v.PublicIP
			mPK = v.PrivateSSHKey
			mPrIP = v.PrivateIP
		}
	}

	utils.SetKey("MASTER_PUBLIC_IP", mIP)
	utils.SetKey("MASTER_PRIVATE_IP", mPrIP)
	utils.SetKey("MASTER_PRIVATE_KEY", base64.StdEncoding.EncodeToString([]byte(mPK)))
	utils.SetKey("MINION_IPS", strings.Join(miIP, ","))
}
