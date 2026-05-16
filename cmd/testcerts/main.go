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
		caKeyFlag = flag.String("cakey", "", "TLS CA certificate key")
		hostFlag = flag.String("host", "localhost", "TLS certificate host for self-signed")
		hostsFlag = flag.String("hosts", "", "List of TLS certificate hosts, separated by comma, if empty: client certificate")
		cnFlag = flag.String("cn", "CA", "Common name of the TLS CA")
	)
	cf := cli.CommonFlags()
	flag.Parse()
	lgr, err := common.CliLogger("vdasync", *cf.LogLevelFlag, *cf.LogFlag)
	if err != nil {
		common.Fatal(lgr, err)
	}
	if *cf.CaCertFlag != "" && *cf.CertFlag == "" {
		// bin/testcerts -ca /tmp/ca-cert.pem -cakey /tmp/ca-key.pem
		if err := tls.NewCaCertFiles(*cf.CaCertFlag, *caKeyFlag, *cnFlag); err != nil {
			common.Fatal(lgr, err)
		}
		return
	}
	if *cf.CertFlag != "" && *cf.CaCertFlag == "" {
		// bin/testcerts -cert /tmp/loc-cert.pem -key /tmp/loc-key.pem
		if err := tls.SelfSignedFiles(*hostFlag, *cf.CertFlag, *cf.KeyFlag); err != nil {
			common.Fatal(lgr, err)
		}
		return
	}
	// bin/testcerts -ca /tmp/ca-cert.pem -cakey /tmp/ca-key.pem
	var hosts []string
	if *hostsFlag != "" {
		hosts = strings.Split(*hostsFlag, ",")
	}
	lgr.Info("hosts", "hosts", hosts)
	if err := tls.NewCertFiles(hosts, *cf.CaCertFlag, *caKeyFlag, *cf.CertFlag, *cf.KeyFlag); err != nil {
		common.Fatal(lgr, err)
	}
}
