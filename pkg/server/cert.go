package server

// Based on https://golang.org/src/crypto/tls/generate_cert.go

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"github.com/pkg/errors"
	"math/big"
	"net"
	"strings"
	"time"
)

type KeyAndCertOpts struct {
	// Host is a comma-separated hostnames and IPs to generate a certificate for.
	Host string
	// ValidFrom is the creation date formatted as Jan 1 15:04:05 2011.
	ValidFrom string
	// ValidFor is the duration that certificate is valid for.
	ValidFor time.Duration
	// IsCA sets whether this cert should be its own Certificate Authority.
	IsCA bool
	// RSABits is the zize of RSA key to generate. Ignored if ecdsa-curve is set.
	RSABits int
	// ECDSACurve is the ECDSA curve to use to generate a key. Valid values are P224, P256 (recommended), P384, P521.
	ECDSACurve string
	// ed25519Key sets whether to generate an Ed25519 key.
	Ed25519Key bool
}

type KeyAndCert struct {
	Key  []byte
	Cert []byte
}

func GenerateKeyAndCert(opts KeyAndCertOpts) (*KeyAndCert, error) {
	// Check required values.
	if len(opts.Host) == 0 {
		return nil, errors.New("Missing required host")
	}

	// Set default values.
	if opts.ValidFor == 0 {
		opts.ValidFor = 365 * 24 * time.Hour
	}
	if opts.RSABits == 0 {
		opts.RSABits = 2048
	}
	if opts.ECDSACurve == "" {
		opts.ECDSACurve = "P256"
	}

	var priv interface{}
	var err error
	switch opts.ECDSACurve {
	case "":
		if opts.Ed25519Key {
			_, priv, err = ed25519.GenerateKey(rand.Reader)
		} else {
			priv, err = rsa.GenerateKey(rand.Reader, opts.RSABits)
		}
	case "P224":
		priv, err = ecdsa.GenerateKey(elliptic.P224(), rand.Reader)
	case "P256":
		priv, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	case "P384":
		priv, err = ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	case "P521":
		priv, err = ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	default:
		return nil, errors.Errorf("Unrecognized elliptic curve: %q", opts.ECDSACurve)
	}
	if err != nil {
		return nil, errors.Wrap(err,"Failed to generate private key")
	}

	var notBefore time.Time
	if len(opts.ValidFrom) == 0 {
		notBefore = time.Now()
	} else {
		notBefore, err = time.Parse("Jan 2 15:04:05 2006", opts.ValidFrom)
		if err != nil {
			return nil, errors.Wrap(err,"Failed to parse creation date")
		}
	}

	notAfter := notBefore.Add(opts.ValidFor)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, errors.Wrap(err,"Failed to generate serial number")
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Acme Co"},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	hosts := strings.Split(opts.Host, ",")
	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}

	if opts.IsCA {
		template.IsCA = true
		template.KeyUsage |= x509.KeyUsageCertSign
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, publicKey(priv), priv)
	if err != nil {
		return nil, errors.Wrap(err,"Failed to create certificate")
	}

	certOut := new(bytes.Buffer)
	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return nil, errors.Wrap(err,"Failed to create certificate")
	}

	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return nil, errors.Wrap(err,"Unable to marshal private key")
	}
	keyOut := new(bytes.Buffer)
	if err := pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes}); err != nil {
		return nil, errors.Wrap(err,"Failed to create key")
	}

	return &KeyAndCert{Key: keyOut.Bytes(), Cert: certOut.Bytes()}, nil
}

func publicKey(priv interface{}) interface{} {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &k.PublicKey
	case *ecdsa.PrivateKey:
		return &k.PublicKey
	case ed25519.PrivateKey:
		return k.Public().(ed25519.PublicKey)
	default:
		return nil
	}
}
