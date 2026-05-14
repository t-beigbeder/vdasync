package tls

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"strings"
	"time"

	"github.com/t-beigbeder/vdasync/internal/common"
)

// SelfSigned generates a new self-signed TLS certificate key pair for the given host
func SelfSigned(host string) (*tls.Certificate, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	keyUsage := x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment
	notBefore := time.Now()
	notAfter := notBefore.Add(365 * 24 * time.Hour)
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, err
	}
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"otvl"},
			CommonName:   "self-signed",
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              keyUsage,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}
	template.DNSNames = append(template.DNSNames, host)
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return nil, err
	}
	certPem := bytes.Buffer{}
	err = pem.Encode(&certPem, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	if err != nil {
		return nil, err
	}
	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return nil, err
	}
	keyPem := bytes.Buffer{}
	if err := pem.Encode(&keyPem, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes}); err != nil {
		return nil, err
	}
	cert, err := tls.X509KeyPair(certPem.Bytes(), keyPem.Bytes())
	return &cert, err
}

// SelfSignedFiles generates a new self-signed TLS certificate key pair files for the given host
func SelfSignedFiles(host string, certFile, keyFile string) error {
	cert, err := SelfSigned(host)
	if err != nil {
		return err
	}
	certPEM := new(bytes.Buffer)
	err = pem.Encode(certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Leaf.Raw,
	})
	if err != nil {
		return err
	}
	err = common.WriteFile(certFile, certPEM.Bytes())
	if err != nil {
		return err
	}

	certPrivKeyPEM := new(bytes.Buffer)
	certPrik, ok := cert.PrivateKey.(*rsa.PrivateKey)
	if !ok {
		return fmt.Errorf("expected RSA private key, got %T", cert.PrivateKey)
	}
	err = pem.Encode(certPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(certPrik),
	})
	if err != nil {
		return err
	}
	err = common.WriteFile(keyFile, certPrivKeyPEM.Bytes())
	if err != nil {
		return err
	}
	return nil
}

var snCount = 1

func getRandSN() (*big.Int, error) {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, err
	}
	return serialNumber, nil
}

// NewCaCert generates a private CA for tests
func NewCaCert() (*x509.CertPool, *x509.Certificate, *rsa.PrivateKey, error) {
	notBefore := time.Now()
	notAfter := notBefore.Add(365 * 24 * time.Hour)
	serialNumber, err := getRandSN()
	if err != nil {
		return nil, nil, nil, err
	}
	caCert := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"otvl"},
			CommonName:   "CA",
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}
	caPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, nil, err
	}
	caBytes, err := x509.CreateCertificate(rand.Reader, caCert, caCert, &caPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return nil, nil, nil, err
	}
	caCert, err = x509.ParseCertificate(caBytes)
	if err != nil {
		return nil, nil, nil, err
	}
	caPEM := new(bytes.Buffer)
	pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})
	certpool := x509.NewCertPool()
	certpool.AppendCertsFromPEM(caPEM.Bytes())
	return certpool, caCert, caPrivKey, nil
}

// NewCaCertFiles generate a private CA PEM files for tests
func NewCaCertFiles(certFile, keyFile string) error {
	_, caCert, caPrik, err := NewCaCert()
	if err != nil {
		return err
	}
	caPEM := new(bytes.Buffer)
	err = pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caCert.Raw,
	})
	if err != nil {
		return err
	}
	err = common.WriteFile(certFile, caPEM.Bytes())
	if err != nil {
		return err
	}
	caPrivKeyPEM := new(bytes.Buffer)
	err = pem.Encode(caPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(caPrik),
	})
	if err != nil {
		return err
	}
	err = common.WriteFile(keyFile, caPrivKeyPEM.Bytes())
	if err != nil {
		return err
	}
	return nil
}

// NewCert generates a client or server (given hosts) certificate from the given CA
func NewCert(hosts []string, caCert *x509.Certificate, caPrivKey *rsa.PrivateKey) (*tls.Certificate, error) {
	notBefore := time.Now()
	notAfter := notBefore.Add(365 * 24 * time.Hour)
	serialNumber, err := getRandSN()
	if err != nil {
		return nil, err
	}
	cn := ""
	if hosts != nil {
		cn = strings.Join(hosts, ",")
	}
	cert := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"otvl"},
			CommonName:   cn,
		},
		IPAddresses: []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		NotBefore:   notBefore,
		NotAfter:    notAfter,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature,
	}
	if len(hosts) > 0 {
		for _, h := range hosts {
			if ip := net.ParseIP(h); ip != nil {
				cert.IPAddresses = append(cert.IPAddresses, ip)
			} else {
				cert.DNSNames = append(cert.DNSNames, h)
			}
		}
		cert.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}
	} else {
		cert.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}
	}
	certPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, err
	}
	certBytes, err := x509.CreateCertificate(rand.Reader, cert, caCert, &certPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return nil, err
	}
	certPEM := new(bytes.Buffer)
	pem.Encode(certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})
	certPrivKeyPEM := new(bytes.Buffer)
	pem.Encode(certPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(certPrivKey),
	})
	tlsCert, err := tls.X509KeyPair(certPEM.Bytes(), certPrivKeyPEM.Bytes())
	if err != nil {
		return nil, err
	}
	return &tlsCert, nil
}

// NewCertFiles generates a client or server (given hosts) certificate files from the given CA
func NewCertFiles(hosts []string, caCertFile, caKeyFile, certFile, keyFile string) error {
	pair, err := tls.LoadX509KeyPair(caCertFile, caKeyFile)
	if err != nil {
		return err
	}
	caPrik, ok := pair.PrivateKey.(*rsa.PrivateKey)
	if !ok {
		return fmt.Errorf("expected RSA private key, got %T", pair.PrivateKey)
	}
	cert, err := NewCert(hosts, pair.Leaf, caPrik)
	if err != nil {
		return err
	}
	certPEM := new(bytes.Buffer)
	err = pem.Encode(certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Leaf.Raw,
	})
	if err != nil {
		return err
	}
	err = common.WriteFile(certFile, certPEM.Bytes())
	if err != nil {
		return err
	}

	certPrivKeyPEM := new(bytes.Buffer)
	certPrik, ok := cert.PrivateKey.(*rsa.PrivateKey)
	if !ok {
		return fmt.Errorf("expected RSA private key, got %T", cert.PrivateKey)
	}
	err = pem.Encode(certPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(certPrik),
	})
	if err != nil {
		return err
	}
	err = common.WriteFile(keyFile, certPrivKeyPEM.Bytes())
	if err != nil {
		return err
	}
	return nil

}

func GetInsecureSkipVerifyConfig() *tls.Config {
	return &tls.Config{
		InsecureSkipVerify: true,
	}
}
