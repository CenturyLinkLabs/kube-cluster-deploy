package deploy

import (
    "os"
    "errors"
    "github.com/mitchellh/goamz/aws"
    "strings"
    "github.com/mitchellh/goamz/ec2"
    "github.com/CenturylinkLabs/kube-cluster-deploy/utils"
    "math/rand"
    "time"
    "fmt"
    "golang.org/x/crypto/ssh"
    "bytes")

type Amazon struct {
    VMSize string
    Location string
    ApiKeyID string
    ApiAccessKey string
    AmiName string
    AmiOwnerId string
    SSHKeyName string
    PrivateKey string
    amzClient *ec2.EC2
    ServerCount int
    TCPOpenPorts   []int
}

func (amz *Amazon)DeployVMs() ([]CloudServer, error) {

    e := amz.init()
    if e != nil {
        return nil, e
    }

    amz.AmiName, e = amz.getAmiID()

    if amz.AmiName == "" || e != nil {
        fmt.Printf("AMI Not found")
        return nil, errors.New("AMI Not found for provisioning. Cannot proceed.")
    }

    g, e := amz.createFWRules()

    if e != nil {
        return nil, e
    }

    req := &ec2.RunInstances{
        ImageId:      amz.AmiName,
        InstanceType: amz.VMSize,
        MinCount:     amz.ServerCount,
        MaxCount:     amz.ServerCount,
        KeyName:      amz.SSHKeyName,
        SecurityGroups: []ec2.SecurityGroup{g},
    }

    resp, e := amz.amzClient.RunInstances(req)
    if e != nil {
        return nil, e
    }

    utils.LogInfo("\nWaiting for servers to provision....")

    var servers []CloudServer
    for _, inst := range resp.Instances {
        s, e := amz.waitForServer(inst)
        if e != nil {
            return servers, e
        }
        s.PrivateSSHKey = amz.PrivateKey
        servers = append(servers, s)
    }
    utils.LogInfo("\nServer provisioning complete...")

    return servers, nil
}

func (amz *Amazon) getAmiID() (string, error) {
    f := ec2.NewFilter()
    f.Add("name","*"+ amz.AmiName +"*")
    f.Add("owner-id",amz.AmiOwnerId)
    im, _ := amz.amzClient.Images(nil, f)
    if im != nil && len(im.Images) >0 {
        return im.Images[0].Id, nil
    }
    return "", errors.New("Image not found")
}

func (amz *Amazon) createFWRules() (ec2.SecurityGroup, error) {

    g := ec2.SecurityGroup{}
    g.Name = "pmx-security-group-" + amz.randSeq(4)
    g.Description = "panamax security group"
    var ps []ec2.IPPerm

    amz.TCPOpenPorts = append(amz.TCPOpenPorts, 22)
    for _, p := range amz.TCPOpenPorts {
        ps = append(ps, ec2.IPPerm{Protocol: "tcp", SourceIPs: []string{"0.0.0.0/0"}, ToPort: p, FromPort: p,})
    }

    _, e := amz.amzClient.CreateSecurityGroup(g)
    if e != nil {
        return ec2.SecurityGroup{}, e
    }
    _, e = amz.amzClient.AuthorizeSecurityGroup(g, ps)
    if e != nil {
        return  ec2.SecurityGroup{}, e
    }
    return g, nil
}

func (amz *Amazon) waitForServer(inst ec2.Instance) (CloudServer, error) {
    for {
        if inst.State.Code == 16 {
            break
        }
        time.Sleep(10 * time.Second)
        resp, e := amz.amzClient.Instances([]string{inst.InstanceId}, &ec2.Filter{})
        if e != nil {
            return CloudServer{}, e
        }
        inst = resp.Reservations[0].Instances[0]
    }
    utils.LogInfo(fmt.Sprintf("\nServer Provisioned: Public IP: %s, Private IP: %s", inst.PublicIpAddress, inst.PrivateIpAddress))
    return CloudServer{PublicIP: inst.PublicIpAddress, Name: inst.DNSName, PrivateIP: inst.PrivateIpAddress}, nil
}

func (amz *Amazon) init() error {

    if amz.ApiKeyID == "" || amz.ApiAccessKey == "" || amz.Location == "" || amz.VMSize == "" {
        return errors.New("\n\nMissing Params Or No Matching AMI found...Check Docs...\n\n")
    }

    os.Setenv("AWS_ACCESS_KEY_ID", amz.ApiKeyID)
    os.Setenv("AWS_SECRET_ACCESS_KEY", amz.ApiAccessKey)

    auth, e := aws.EnvAuth()
    if e != nil {
        return e
    }

    var r aws.Region
    for _, r = range aws.Regions {
        if strings.Contains(amz.Location, r.Name) {
            break
        }
    }

    amz.amzClient = ec2.New(auth, r)
    return nil
}


func (amz *Amazon) ImportKey(pk string) (string, error) {

    e := amz.init()
    if e != nil {
        return "", e
    }

    kn := "pmx-installer-" + amz.randSeq(4)
    _, e = amz.amzClient.ImportKeyPair(kn, pk)
    if e != nil {
        return "", e
    }
    return kn, nil
}

func (amz Amazon)randSeq(n int) string {
    var letters = []rune("abcdefghijklmnopqrstuvwxyz")
    rand.Seed( time.Now().UTC().UnixNano())
    b := make([]rune, n)
    for i := range b {
        b[i] = letters[rand.Intn(len(letters))]
    }
    return string(b)
}

func(amz Amazon) ExecSSHCmd(publicIP string, privateKey string, command string) string {

    utils.LogInfo("\nWaiting for server to start before adding ssh keys")
    e := utils.WaitForSSH(publicIP)

    if e != nil {
        fmt.Printf("Panic")
        panic(e)
    }

    k, e := ssh.ParsePrivateKey([]byte(privateKey))
    fmt.Printf("Key read")
    if e != nil {
        fmt.Println(e)
        panic(e)
    }

    c := &ssh.ClientConfig{
        User: "ec2-user",
        Auth: []ssh.AuthMethod{ssh.PublicKeys(k),},
    }

    cn, _ := ssh.Dial("tcp", publicIP+":22", c)
    s, _ := cn.NewSession()
    defer s.Close()

    var sOut bytes.Buffer
    s.Stdout = &sOut
    s.Run(command)

    utils.LogInfo(sOut.String())
    utils.LogInfo("\nCommand Complete")

    return sOut.String()
}