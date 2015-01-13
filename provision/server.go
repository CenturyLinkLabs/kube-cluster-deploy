package provision

type Server struct {
	Name          string
	PublicIP      string
	PrivateIP     string
	PublicSSHKey  string
	PrivateSSHKey string
}
