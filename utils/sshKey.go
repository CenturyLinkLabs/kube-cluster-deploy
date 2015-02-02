package utils

import (
	"code.google.com/p/go.crypto/ssh"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
)

// CreateSSHKey is used to generate rsa private & public keys that are used to
// set keys for ssh login to a server.
func CreateSSHKey() (string, string, error) {

	pk, e := rsa.GenerateKey(rand.Reader, 2014)
	if e != nil {
		return "", "", e
	}

	pkDer := x509.MarshalPKCS1PrivateKey(pk)
	pkBlk := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   pkDer,
	}
	pkPem := string(pem.EncodeToMemory(&pkBlk))

	pubK := pk.PublicKey

	pub, e := ssh.NewPublicKey(&pubK)
	if e != nil {
		return "", "", e
	}

	pubBytes := ssh.MarshalAuthorizedKey(pub)

	return string(pkPem), string(pubBytes), nil
}
