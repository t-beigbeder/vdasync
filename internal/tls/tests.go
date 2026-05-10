package tls

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/t-beigbeder/vdasync/internal/common"
)

func getCertDirs(testDir string, hosts []string) (string, string, string, string) {
	if os.Getenv("QSTF_TEST_CACHE") == "" {
		return testDir, testDir, testDir, testDir
	}
	tcd := os.Getenv("QSTF_TEST_CERTS_DIR")
	if tcd == "" {
		tcd = path.Join(os.TempDir(), "qstf_test_certs")
	}
	svh := fmt.Sprintf("%064x", sha256.Sum256([]byte(strings.Join(hosts, ","))))
	return path.Join(tcd, "ca"),
		path.Join(tcd, svh),
		path.Join(tcd, "c1"),
		path.Join(tcd, "c2")
}

func makeCertIf(certDir string, maker func() error) error {
	if os.Getenv("QSTF_TEST_CACHE") == "" {
		return maker()
	}
	ffn := path.Join(certDir, "done.flag")
	if common.FileExists(ffn) {
		return nil
	}
	if !common.FileExists(certDir) {
		if err := os.MkdirAll(certDir, 0700); err != nil {
			return err
		}
	}
	if err := maker(); err != nil {
		return err
	}
	if err := common.WriteFile(ffn, nil); err != nil {
		return err
	}
	return nil
}

func NewTestCerts(testDir string, hosts []string, hasSecClient bool) (map[string]string, error) {
	caDir, svDir, c1Dir, c2Dir := getCertDirs(testDir, hosts)
	cfs := map[string]string{
		"cac": path.Join(caDir, "cacert.pem"),
		"cak": path.Join(caDir, "cacert-key.pem"),
		"svc": path.Join(svDir, "svcert.pem"),
		"svk": path.Join(svDir, "svcert-key.pem"),
		"c1c": path.Join(c1Dir, "c1cert.pem"),
		"c1k": path.Join(c1Dir, "c1cert-key.pem"),
	}
	if err := makeCertIf(caDir, func() error {
		return NewCaCertFiles(cfs["cac"], cfs["cak"])
	}); err != nil {
		return nil, err
	}
	if err := makeCertIf(svDir, func() error {
		return NewCertFiles(hosts, cfs["cac"], cfs["cak"], cfs["svc"], cfs["svk"])
	}); err != nil {
		return nil, err
	}
	if err := makeCertIf(c1Dir, func() error {
		return NewCertFiles(nil, cfs["cac"], cfs["cak"], cfs["c1c"], cfs["c1k"])
	}); err != nil {
		return nil, err
	}
	if !hasSecClient {
		return cfs, nil
	}
	cfs["c2c"] = path.Join(c2Dir, "c2cert.pem")
	cfs["c2k"] = path.Join(c2Dir, "c2cert-key.pem")
	if err := makeCertIf(c2Dir, func() error {
		return NewCertFiles(nil, cfs["cac"], cfs["cak"], cfs["c2c"], cfs["c2k"])
	}); err != nil {
		return nil, err
	}
	return cfs, nil
}
