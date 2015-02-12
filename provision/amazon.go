package provision

import (
    "os"
    "github.com/CenturylinkLabs/kube-cluster-deploy/utils"
    "errors"
    "github.com/CenturylinkLabs/kube-cluster-deploy/deploy"
    "fmt"
    "encoding/base64"
    "strings"
    "strconv")

type Amazon struct {

}

// NewAmazon is used to create a new client for using Amazon client to
// create RHEL 7 server cluster.
func NewAmazon() *Amazon {
    cl := new(Amazon)
    return cl
}

func (amz *Amazon) ProvisionCluster() ([]deploy.CloudServer, error) {

    utils.LogInfo("\nProvisioning cluster in Amazon EC2")

    snMaster := "Master"
    snMinion := "Minion"

    apiID := os.Getenv("AWS_ACCESS_KEY_ID")
    apiK := os.Getenv("AWS_SECRET_ACCESS_KEY")
    loc := os.Getenv("REGION")
    vmSize := os.Getenv("VM_SIZE")
    cnt, e := strconv.Atoi(os.Getenv("MINION_COUNT"))

    if apiID == "" || apiK == "" || loc == "" || vmSize == "" {
        return nil, errors.New("\n\nMissing Params Or No Matching AMI found...Check Docs...\n\n")
    }

    pk, puk, _ := utils.CreateSSHKey()

    c := &deploy.Amazon{}
    c.ApiAccessKey = apiK
    c.ApiKeyID = apiID
    c.Location = loc
    c.PrivateKey = pk
    c.PublicKey = puk
    c.ServerCount = cnt + 1
    c.VMSize = vmSize
    c.AmiName = "RHEL-7.0_HVM_GA"
    c.AmiOwnerId = "309956199498"

    c.TCPOpenPorts = []int{8080, 4001, 7001, 10250}
    for _, p := range strings.Split(os.Getenv("OPEN_TCP_PORTS"), ",") {
        v, e := strconv.Atoi(p)
        if e == nil {
            c.TCPOpenPorts = append(c.TCPOpenPorts, v)
        }
    }

    c.ServerNames = append(c.ServerNames, snMaster)
    for i := 0; i < c.ServerCount - 1 ; i++ {
        c.ServerNames = append(c.ServerNames, fmt.Sprintf("%s-%d", snMinion, i))
    }

    servers, e := c.DeployVMs()

    if e != nil {
        return nil, e
    }

    for _, s := range servers {
        s.PublicSSHKey = puk
        if s.Name != snMaster {
            s.PrivateSSHKey = ""
        } else {
            cmd := fmt.Sprintf("echo -e \"%s\" >> ~/.ssh/id_rsa && chmod 400 ~/.ssh/id_rsa", s.PrivateSSHKey)
            c.ExecSSHCmd(s.PublicIP, s.PrivateSSHKey, cmd)
        }
    }

    utils.LogInfo("\nCluster Creating Complete...")
    utils.SetKey("AMAZON_SSH_KEY_NAME", c.SSHKeyName)
    utils.SetKey("MASTER_PUBLIC_KEY", base64.StdEncoding.EncodeToString([]byte(puk)))
    utils.SetKey("UBUNTU_LOGIN_USER","ubuntu")
    utils.SetKey("RHEL_LOGIN_USER","ec2-user")

    return servers, nil
}
