package main

import (
	"flag"
	"strings"

	"github.com/t-beigbeder/vdasync/internal/cli"
	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/internal/tls"
)

func main() {
	var (
		CaKeyFlag = flag.String("cakey", "", "TLS CA certificate key")
		HostFlag = flag.String("host", "localhost", "TLS certificate host for self-signed")
		HostsFlag = flag.String("hosts", "", "List of TLS certificate hosts, separated by comma, if empty: client certificate")
	)
	cf := cli.CommonFlags()
	flag.Parse()
	lgr, err := common.CliLogger("vdasync", *cf.LogLevelFlag, *cf.LogFlag)
	if err != nil {
		common.Fatal(lgr, err)
	}
	if *cf.CaCertFlag != "" && *cf.CertFlag == "" {
		// bin/testcerts -ca /tmp/ca-cert.pem -cakey /tmp/ca-key.pem
		if err := tls.NewCaCertFiles(*cf.CaCertFlag, *CaKeyFlag); err != nil {
			common.Fatal(lgr, err)
		}
		return
	}
	if *cf.CertFlag != "" && *cf.CaCertFlag == "" {
		// bin/testcerts -cert /tmp/loc-cert.pem -key /tmp/loc-key.pem
		if err := tls.SelfSignedFiles(*HostFlag, *cf.CertFlag, *cf.KeyFlag); err != nil {
			common.Fatal(lgr, err)
		}
		return
	}
	// bin/testcerts -ca /tmp/ca-cert.pem -cakey /tmp/ca-key.pem
	var hosts []string
	if *HostsFlag != "" {
		hosts = strings.Split(*HostsFlag, ",")
	}
	lgr.Info("hosts", "hosts", hosts)
	if err := tls.NewCertFiles(hosts, *cf.CaCertFlag, *CaKeyFlag, *cf.CertFlag, *cf.KeyFlag); err != nil {
		common.Fatal(lgr, err)
	}
}
