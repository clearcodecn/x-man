package src

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"github.com/clearcodecn/xman/cert"
	"github.com/hashicorp/golang-lru"
	errors "github.com/pkg/errors"
	"math/big"
	"net"
	"time"
)

var (
	defaultCertManager, _ = newCertManager(cert.CertFrom)
)

type CertFrom func() (cert []byte, key []byte)

type certManager struct {
	rootCa  *x509.Certificate
	rootKey interface{}
	cache   *lru.Cache

	rootRaw []byte
	keyRaw  []byte
}

func newCertManager(cf CertFrom) (*certManager, error) {
	if cf == nil {
		cf = cert.CertFrom
	}
	cm := new(certManager)
	ca, key := cf()
	cm.rootRaw = ca
	cm.keyRaw = key
	
	block, _ := pem.Decode(ca)
	rootCa, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, err
	}
	cm.rootCa = rootCa
	block, _ = pem.Decode(key)
	rootKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	cm.rootKey = rootKey
	cm.cache, _ = lru.New(1024)
	return cm, nil
}

func (cm *certManager) GenerateTlsByHost(host string) (*tls.Config, error) {
	if h, _, _ := net.SplitHostPort(host); host != "" {
		host = h
	}

	if v, ok := cm.cache.Get(host); ok {
		return v.(*tls.Config), nil
	}
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{`X-Man`},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().AddDate(1, 0, 0),

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}
	if ip := net.ParseIP(host); err != nil {
		template.IPAddresses = append(template.IPAddresses, ip)
	} else {
		template.DNSNames = append(template.DNSNames, host)
	}
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, cm.rootCa, &priv.PublicKey, cm.rootKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create certificate")
	}
	certBlock := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: derBytes,
	}
	serverCert := pem.EncodeToMemory(certBlock)

	keyBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(priv),
	}
	serverKey := pem.EncodeToMemory(keyBlock)

	conf, err := tls.X509KeyPair(serverCert, serverKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load x509 key pair")
	}

	tlsConf := &tls.Config{
		Certificates: []tls.Certificate{conf},
	}

	cm.cache.Add(host, tlsConf)

	return tlsConf, nil
}

func (cm *certManager) RootRaw() []byte {
	return cm.rootRaw
}
