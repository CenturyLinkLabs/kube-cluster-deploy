package provision

import (
	"code.google.com/p/go.crypto/ssh"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
)

func createSSHKey() (string, string) {

	privateKey, err := rsa.GenerateKey(rand.Reader, 2014)
	if err != nil {
		panic(err)
	}

	privateKeyDer := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyBlock := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   privateKeyDer,
	}
	privateKeyPem := string(pem.EncodeToMemory(&privateKeyBlock))

	publicKey := privateKey.PublicKey

	pub, err := ssh.NewPublicKey(&publicKey)
	if err != nil {
		panic(err)
	}

	pubBytes := ssh.MarshalAuthorizedKey(pub)

	//ioutil.WriteFile("key", []byte(privateKeyPem), 0777)
	//ioutil.WriteFile("key.pub", []byte(pubBytes), 0777)

	return string(privateKeyPem), string(pubBytes)
}
