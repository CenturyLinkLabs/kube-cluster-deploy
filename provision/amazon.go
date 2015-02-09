package provision

import (
    "os"
    "github.com/CenturylinkLabs/kube-cluster-deploy/utils"
    "errors"
    "github.com/CenturylinkLabs/kube-cluster-deploy/deploy"
    "fmt")

type Amazon struct {

}

//AKIAIJAZWSCQEU32IMPQ
//Secret Access Key:
//eUdpKqdcXEItsjGzeVFmqtB/SIGbvZQ78G+S5nhV
// NewAmazon is used to create a new client for using Amazon client to
// create RHEL 7 server cluster.
func NewAmazon() *Amazon {
    cl := new(Amazon)
    return cl
}

func (amz *Amazon) ProvisionCluster(params Params) ([]deploy.CloudServer, error) {

    utils.LogInfo("\nProvisioning cluster in Amazon EC2")

    apiID := os.Getenv("AWS_ACCESS_KEY_ID")
    apiK := os.Getenv("AWS_SECRET_ACCESS_KEY")
    location := os.Getenv("REGION")
    vmSize := os.Getenv("VM_SIZE")

    if apiID == "" || apiK == "" || location == "" || vmSize == "" {
        return nil, errors.New("\n\nMissing Params Or No Matching AMI found...Check Docs...\n\n")
    }

    pk, puk, _ := utils.CreateSSHKey()

    c := &deploy.Amazon{}
    c.ApiAccessKey = apiK
    c.ApiKeyID = apiID
    c.Location = location
    c.PrivateKey = pk
    c.ServerCount = params.MinionCount + 1
    c.VMSize = vmSize
    c.TCPOpenPorts = []int{8080}

    kn, e := c.ImportKey(puk)
    if e != nil {
        return nil, e
    }
    c.SSHKeyName = kn

    servers, e := c.DeployVMs()
    if e != nil {
        return nil, e
    }

    for i, s := range servers {
        s.PublicSSHKey = puk
        if i > 0 {
            s.PrivateSSHKey = ""
        }
    }

    cmd := fmt.Sprintf("echo -e \"%s\" >> ~/.ssh/id_rsa && chmod 400 ~/.ssh/id_rsa", servers[0].PrivateSSHKey)
    c.ExecSSHCmd(servers[0].PublicIP, servers[0].PrivateSSHKey, cmd)

    utils.LogInfo("\nCluster Creating Complete...")

    return servers, nil
}
