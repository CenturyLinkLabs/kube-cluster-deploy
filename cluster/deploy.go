package main

import (
	"encoding/base64"
	"github.com/CenturylinkLabs/kube-cluster-deploy/provision"
	"github.com/CenturylinkLabs/kube-cluster-deploy/utils"
	"os"
	"strconv"
	"strings"
    "fmt"
    )

func main() {
    fmt.Printf("Starting cluster deployment")

	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r)
			os.Exit(1)
		}
	}()

    c, e := strconv.Atoi(os.Getenv("MINION_COUNT"))
	if c == 0 || e != nil {
		panic("\nPlease make sure you have at least one minion in the cluster.")
	}

    cfg, e := utils.LoadJsonConfig()

    p := cfg["provider"]
	cp := provision.New(p)
	s, e := cp.ProvisionCluster()
	if e != nil {
		panic(e.Error())
	}

    mPuIP := ""
    mPrIP := ""
    mPK := ""

    var miIP []string

    for _, v := range s {
		if v.PrivateSSHKey == "" || mPuIP != "" {
			miIP = append(miIP, v.PrivateIP)
		} else {
			mPuIP = v.PublicIP
            mPrIP = v.PrivateIP
            mPK = v.PrivateSSHKey
		}
	}

    utils.SetKey("CLOUD_PROVIDER",p)
	utils.SetKey("MASTER_PUBLIC_IP", mPuIP)
	utils.SetKey("MASTER_PRIVATE_IP", mPrIP)
	utils.SetKey("MASTER_PRIVATE_KEY", base64.StdEncoding.EncodeToString([]byte(mPK)))
	utils.SetKey("MINION_IPS", strings.Join(miIP, ","))
}
