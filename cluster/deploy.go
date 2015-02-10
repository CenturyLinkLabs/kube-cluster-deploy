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

    c, _ := strconv.Atoi(os.Getenv("MINION_COUNT"))

	if c == 0 {
		panic("\nPlease make sure you have at least one minion in the cluster.")
	}

	params := provision.Params{MinionCount: c}
    p := "amazon"

	cp := provision.New(p)
	s, e := cp.ProvisionCluster(params)

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

	utils.SetKey("MASTER_PUBLIC_IP", mPuIP)
	utils.SetKey("MASTER_PRIVATE_IP", mPrIP)
	utils.SetKey("MASTER_PRIVATE_KEY", base64.StdEncoding.EncodeToString([]byte(mPK)))
	utils.SetKey("MINION_IPS", strings.Join(miIP, ","))
    utils.SetKey("UBUNTU_LOGIN_USER","ubuntu")
    utils.SetKey("RHEL_LOGIN_USER","ec2-user")
}
