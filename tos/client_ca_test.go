package tos

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io/ioutil"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func generateValidCACert(t *testing.T) []byte {
	// Generate a new private key
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	// Create a template for the certificate
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Test Org"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour * 24),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	// Create a self-signed certificate
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		t.Fatal(err)
	}

	// Encode the certificate to PEM format
	pemBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: derBytes,
	})

	return pemBytes
}

func TestNewClient_WithValidCaCrt(t *testing.T) {
	// 1. Generate a valid CA certificate
	caContent := generateValidCACert(t)

	// 2. Create a temporary file
	tmpfile, err := ioutil.TempFile("", "valid_ca_*.crt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name()) // clean up

	if _, err := tmpfile.Write(caContent); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	// 3. Attempt to create a client with the valid CA certificate
	client, err := NewClient("tos-cn-beijing.volces.com", WithCaCrt(tmpfile.Name()))

	// 4. Assert that no error is returned and client is created
	assert.NoError(t, err)
	assert.NotNil(t, client)
	
	// Optional: verify that the RootCAs pool is not nil (though this is internal)
	// We can't easily access internal fields, but the lack of error implies success.
}

func TestNewClient_WithInvalidCaCrt(t *testing.T) {
	// 1. Create a temporary file with invalid content
	tmpfile, err := ioutil.TempFile("", "invalid_ca_*.crt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name()) // clean up

	if _, err := tmpfile.Write([]byte("INVALID PEM CONTENT")); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	// 2. Attempt to create a client with the invalid CA certificate
	_, err = NewClient("tos-cn-beijing.volces.com", WithCaCrt(tmpfile.Name()))

	// 3. Assert that an error is returned and it contains the specific message
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse CA certificate")
}

func TestNewClient_WithNonExistentCaCrt(t *testing.T) {
	// Test file not found case as well
	_, err := NewClient("tos-cn-beijing.volces.com", WithCaCrt("/path/to/non/existent/file.crt"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read CA certificate")
}

func generateValidClientCertAndKey(t *testing.T) ([]byte, []byte) {
	// Generate a new private key
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	// Create a template for the certificate
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Test Client Org"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour * 24),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}

	// Create a self-signed certificate
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		t.Fatal(err)
	}

	// Encode the certificate to PEM format
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: derBytes,
	})

	// Encode the private key to PEM format
	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(priv),
	})

	return certPEM, keyPEM
}

func TestNewClient_WithValidClientCrt(t *testing.T) {
	certPEM, keyPEM := generateValidClientCertAndKey(t)

	certFile, err := ioutil.TempFile("", "client_cert_*.crt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(certFile.Name())
	if _, err := certFile.Write(certPEM); err != nil {
		t.Fatal(err)
	}
	certFile.Close()

	keyFile, err := ioutil.TempFile("", "client_key_*.key")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(keyFile.Name())
	if _, err := keyFile.Write(keyPEM); err != nil {
		t.Fatal(err)
	}
	keyFile.Close()

	client, err := NewClient("tos-cn-beijing.volces.com", WithClientCrt(certFile.Name(), keyFile.Name()))
	assert.NoError(t, err)
	assert.NotNil(t, client)
}

func TestNewClient_WithInvalidClientCrtPath(t *testing.T) {
	_, err := NewClient("tos-cn-beijing.volces.com", WithClientCrt("/path/to/non/existent.crt", "/path/to/non/existent.key"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load client certificate")
}

func TestNewClient_WithInvalidClientCrtContent(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "invalid_client_*.crt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte("INVALID CONTENT")); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	_, err = NewClient("tos-cn-beijing.volces.com", WithClientCrt(tmpfile.Name(), tmpfile.Name()))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load client certificate")
}
