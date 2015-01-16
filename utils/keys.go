package utils

import (
	"fmt"
	"github.com/coreos/go-etcd/etcd"
	"os"
	"strings"
)

func SetKey(key string, value string) error {
	logTo := os.Getenv("LOG_TO")
	if logTo == "" {
		logTo = "stdout"
	}
	logTo = strings.ToLower(logTo)

	if logTo == "etcd" {
		println("Logging Keys to etcd...")
		var ec *etcd.Client
		eIP := os.Getenv("ETCD_API")
		if eIP == "" {
			eIP = "172.17.42.1:4001"
		}
		eIP = fmt.Sprintf("http://%s", eIP)
		ms := []string{eIP}
		ec = etcd.NewClient(ms)
		_, e := ec.Set(key, value, 0)

		if e != nil {
			return e
		}
	} else {
		fmt.Printf("\n----BEGIN PANAMAX DATA----\n%s=%s\n----END PANAMAX DATA----\n", key, value)
	}
	return nil
}
