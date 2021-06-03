package ssl

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io/ioutil"
	"math/big"
	"net"
	"time"
)

func bigIntHash(n *big.Int) []byte {
	h := sha1.New()
	h.Write(n.Bytes())
	return h.Sum(nil)
}

// GenerateCA generate a certificate authority
func GenerateCA(organization, country string, years int, keylength int) ([]byte, []byte, error) {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, _ := rand.Int(rand.Reader, serialNumberLimit)
	caPrivKey, err := rsa.GenerateKey(rand.Reader, keylength)
	if err != nil {
		return nil, nil, err
	}
	ca := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization:  []string{organization},
			Country:       []string{country},
			Province:      []string{""},
			Locality:      []string{""},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
		},
		SignatureAlgorithm:    x509.SHA256WithRSA,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(years, 0, 0),
		IsCA:                  true,
		SubjectKeyId:          bigIntHash(caPrivKey.N),
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}
	caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &caPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return nil, nil, err
	}
	caPEM := new(bytes.Buffer)
	pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})
	caPrivKeyPEM := new(bytes.Buffer)
	pem.Encode(caPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(caPrivKey),
	})
	return caPEM.Bytes(), caPrivKeyPEM.Bytes(), nil
}

// GenerateClientCertificate generate a client certificate
func GenerateClientCertificate(ca string, cakey string, organization, country string, seconds int, keylength int) ([]byte, []byte, error) {
	var err error
	var cacert *x509.Certificate
	var capriv *rsa.PrivateKey
	cacert, err = PEMToCertificate(ca)
	if err != nil {
		return nil, nil, err
	}
	capriv, err = PEMToPrivateKey(cakey)
	if err != nil {
		return nil, nil, err
	}
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, _ := rand.Int(rand.Reader, serialNumberLimit)
	certPrivKey, err := rsa.GenerateKey(rand.Reader, keylength)
	if err != nil {
		return nil, nil, err
	}
	cert := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization:  []string{organization},
			Country:       []string{country},
			Province:      []string{""},
			Locality:      []string{""},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
		},
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(time.Duration(seconds) * time.Second),
		SubjectKeyId: bigIntHash(certPrivKey.N),
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}
	certBytes, err := x509.CreateCertificate(rand.Reader, cert, cacert, &certPrivKey.PublicKey, capriv)
	if err != nil {
		return nil, nil, err
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
	return certPEM.Bytes(), certPrivKeyPEM.Bytes(), nil
}

// PEMToCertificate convert the pem file to a certificate
func PEMToCertificate(path string) (*x509.Certificate, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(data)
	return x509.ParseCertificate(block.Bytes)
}

// PEMToCertificate convert the pem file to a certificate
func PEMToPrivateKey(path string) (*rsa.PrivateKey, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(data)
	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

// BytesToCertificate convert the pem bytearray to a certificate
func BytesToCertificate(data []byte) (*x509.Certificate, error) {
	block, _ := pem.Decode(data)
	return x509.ParseCertificate(block.Bytes)
}

// BytesToPrivateKey convert the pem bytearray to a certificate
func BytesToPrivateKey(data []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(data)
	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

// BytesToPrivateKey convert the pem bytearray to a certificate
func ChainToCertificate(cert string, key string) (tls.Certificate, error) {
	return tls.LoadX509KeyPair(cert, key)
}
