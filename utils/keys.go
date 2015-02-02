package utils

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/coreos/go-etcd/etcd"
	"os"
	"strings"
)

// SetKey is used to log key value pairs to stdout/etcd so that dray can pass
// them down to subsequent images as needed. By default keys are logged to
// stdout as ----BEGIN PANAMAX DATA----\nkey=value\n----END PANAMAX DATA----
// tags. If LOG_TO env var is set to etcd, then keys are logged to etcd.
// The etcd api is set using ETCD_API env variable.
func SetKey(key string, value string) error {
	logTo := os.Getenv("LOG_TO")
	if logTo == "" {
		logTo = "stdout"
	}
	logTo = strings.ToLower(logTo)

	if logTo == "etcd" {
		log.Info("Logging Keys to etcd...")
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
