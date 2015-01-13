package main

import (
	"fmt"
	"github.com/coreos/go-etcd/etcd"
	"os"
	"strings"
)

func setKey(key string, value string) {
	logTo := os.Getenv("LOG_TO")
	if logTo == "" {
		logTo = "stdout"
	}
	logTo = strings.ToLower(logTo)

	if logTo == "etcd" {
		println("Logging Keys to etcd...")
		var etcdClient *etcd.Client
		etcdIP := os.Getenv("ETCD_API")
		if etcdIP == "" {
			etcdIP = "172.17.42.1:4001"
		}
		etcdIP = fmt.Sprintf("http://%s", etcdIP)
		machines := []string{etcdIP}
		etcdClient = etcd.NewClient(machines)
		_, err := etcdClient.Set(key, value, 0)

		if err != nil {
			fmt.Println(err)
		}
	} else {
		fmt.Printf("\n----BEGIN PANAMXAX DATA----%s=%s----END PANAMAX DATA----", key, value)
	}
}
