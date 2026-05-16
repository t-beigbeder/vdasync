package tls

import (
	"context"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/vdasync/internal/common"
)

func TestSelfSignedKP(t *testing.T) {
	cert, err := SelfSigned("localhost")
	if err != nil {
		t.Fatal(err)
	}
	cfg := &tls.Config{Certificates: []tls.Certificate{*cert}}
	srv := &http.Server{
		Addr:         "localhost:9443",
		TLSConfig:    cfg,
		ReadTimeout:  time.Minute,
		WriteTimeout: time.Minute,
	}
	go func() {
		_ = srv.ListenAndServeTLS("", "")
	}()
	defer func() { _ = srv.Shutdown(context.TODO()) }()
	time.Sleep(200 * time.Millisecond)
	hc := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	get, err := hc.Get("https://localhost:9443")
	if err != nil {
		t.Fatal(err)
	}
	if get.StatusCode != http.StatusNotFound {
		t.Fatalf("status code %d", get.StatusCode)
	}
}

func TestNewServerCert(t *testing.T) {
	pool, caCert, caPrik, err := NewCaCert("CA")
	require.NoError(t, err)
	_, _, _ = pool, caCert, caPrik
	tlsCert, err := NewCert([]string{"0.0.0.0", "localhost"}, caCert, caPrik)
	require.NoError(t, err)
	_ = tlsCert

	cfg := &tls.Config{Certificates: []tls.Certificate{*tlsCert}}
	srv := &http.Server{
		Addr:         "0.0.0.0:9443",
		TLSConfig:    cfg,
		ReadTimeout:  time.Minute,
		WriteTimeout: time.Minute,
	}
	go func() {
		_ = srv.ListenAndServeTLS("", "")
	}()
	defer func() { _ = srv.Shutdown(context.TODO()) }()
	time.Sleep(200 * time.Millisecond)
	hc := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{RootCAs: pool},
		},
	}
	get, err := hc.Get("https://localhost:9443")
	if err != nil {
		t.Fatal(err)
	}
	if get.StatusCode != http.StatusNotFound {
		t.Fatalf("status code %d", get.StatusCode)
	}
}

func TestNewClientServerCert(t *testing.T) {
	pool, caCert, caPrik, err := NewCaCert("CA")
	require.NoError(t, err)
	sCert, err := NewCert([]string{"0.0.0.0", "localhost"}, caCert, caPrik)
	require.NoError(t, err)
	cCert, err := NewCert(nil, caCert, caPrik)
	require.NoError(t, err)

	cfg := &tls.Config{
		Certificates: []tls.Certificate{*sCert},
		ClientCAs:    pool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
	}
	srv := &http.Server{
		Addr:         "0.0.0.0:9443",
		TLSConfig:    cfg,
		ReadTimeout:  time.Minute,
		WriteTimeout: time.Minute,
	}
	go func() {
		_ = srv.ListenAndServeTLS("", "")
	}()
	defer func() { _ = srv.Shutdown(context.TODO()) }()
	time.Sleep(200 * time.Millisecond)
	hc := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				Certificates: []tls.Certificate{*cCert},
				RootCAs:      pool,
			},
		},
	}
	get, err := hc.Get("https://localhost:9443")
	if err != nil {
		t.Fatal(err)
	}
	if get.StatusCode != http.StatusNotFound {
		t.Fatalf("status code %d", get.StatusCode)
	}
}

func TestNewCaCertFiles(t *testing.T) {
	td := t.TempDir()
	err := NewCaCertFiles(
		filepath.Join(td, "cacert.pem"),
		filepath.Join(td, "cacert-key.pem"), "CA")
	require.NoError(t, err)
	pair, err := tls.LoadX509KeyPair(filepath.Join(td, "cacert.pem"), filepath.Join(td, "cacert-key.pem"))
	require.NoError(t, err)
	_ = pair
	caPrik, ok := pair.PrivateKey.(*rsa.PrivateKey)
	require.True(t, ok)
	cert, err := NewCert(nil, pair.Leaf, caPrik)
	require.NoError(t, err)
	_ = cert
	caPEM, err := common.LoadFile(filepath.Join(td, "cacert.pem"))
	require.NoError(t, err)
	certpool := x509.NewCertPool()
	ok = certpool.AppendCertsFromPEM(caPEM)
	require.True(t, ok)
}

func TestNewCertFiles(t *testing.T) {
	td := t.TempDir()
	err := NewCaCertFiles(
		filepath.Join(td, "cacert.pem"),
		filepath.Join(td, "cacert-key.pem"), "CA")
	require.NoError(t, err)
	err = NewCertFiles(nil,
		filepath.Join(td, "cacert.pem"),
		filepath.Join(td, "cacert-key.pem"),
		filepath.Join(td, "cert.pem"),
		filepath.Join(td, "cert-key.pem"),
	)
	require.NoError(t, err)
	pair, err := tls.LoadX509KeyPair(
		filepath.Join(td, "cert.pem"),
		filepath.Join(td, "cert-key.pem"),
	)
	require.NoError(t, err)
	_ = pair
	err = SelfSignedFiles("test-host", filepath.Join(td, "self.pem"), filepath.Join(td, "self-key.pem"))
	require.NoError(t, err)
	pair, err = tls.LoadX509KeyPair(
		filepath.Join(td, "self.pem"),
		filepath.Join(td, "self-key.pem"),
	)
	require.NoError(t, err)
}

func TestNewTestCerts(t *testing.T) {
	if os.Getenv("OTVL_TEST_FULL") == "" {
		t.Skip("OTVL_TEST_FULL not set")
	}
	logger := common.GetLogger()
	os.Setenv("OTVL_TEST_CACHE", "")
	td := t.TempDir()
	logger.Info("TestNewTestCerts", "msg", "first no cache")
	cfs, err := NewTestCerts(td, []string{"0.0.0.0", "localhost"}, false)
	require.NoError(t, err)
	td = t.TempDir()
	logger.Info("TestNewTestCerts", "msg", "second no cache, second client")
	cfs, err = NewTestCerts(td, []string{"0.0.0.0", "localhost"}, true)
	require.NoError(t, err)
	os.Setenv("OTVL_TEST_CACHE", "1")
	td = t.TempDir()
	logger.Info("TestNewTestCerts", "msg", "first in cache")
	cfs, err = NewTestCerts(td, []string{"0.0.0.0", "localhost"}, false)
	require.NoError(t, err)
	td = t.TempDir()
	logger.Info("TestNewTestCerts", "msg", "second in cache")
	cfs, err = NewTestCerts(td, []string{"0.0.0.0", "localhost"}, false)
	require.NoError(t, err)
	td = t.TempDir()
	logger.Info("TestNewTestCerts", "msg", "third in cache, second client")
	cfs, err = NewTestCerts(td, []string{"0.0.0.0", "localhost"}, true)
	require.NoError(t, err)
	td = t.TempDir()
	logger.Info("TestNewTestCerts", "msg", "fourth in cache, new server and second client")
	cfs, err = NewTestCerts(td, []string{"localhost"}, true)
	require.NoError(t, err)
	logger.Info("TestNewTestCerts", "msg", "fourth in cache, new server done")
	_ = cfs
}

func TestNewClientServerCertFiles(t *testing.T) {
	td := t.TempDir()
	cfs, err := NewTestCerts(td, []string{"0.0.0.0", "localhost"}, false)
	require.NoError(t, err)
	sCfg, err := GetMTlsServerConfig(cfs["cac"], cfs["svc"], cfs["svk"])
	require.NoError(t, err)
	srv := &http.Server{
		Addr:         "0.0.0.0:9443",
		TLSConfig:    sCfg,
		ReadTimeout:  time.Minute,
		WriteTimeout: time.Minute,
	}
	go func() {
		_ = srv.ListenAndServeTLS("", "")
	}()
	defer func() { _ = srv.Shutdown(context.TODO()) }()
	time.Sleep(200 * time.Millisecond)

	cCfg, err := GetMTlsClientConfig(cfs["cac"], cfs["c1c"], cfs["c1k"])
	require.NoError(t, err)
	hc := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: cCfg,
		},
	}
	get, err := hc.Get("https://localhost:9443")
	if err != nil {
		t.Fatal(err)
	}
	if get.StatusCode != http.StatusNotFound {
		t.Fatalf("status code %d", get.StatusCode)
	}
}
