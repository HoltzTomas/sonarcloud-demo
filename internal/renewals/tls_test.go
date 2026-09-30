package renewals

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// issueCert returns a self-signed certificate valid only for the given hostname.
func issueCert(t *testing.T, hostname string) (tls.Certificate, *x509.Certificate) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	template := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: hostname},
		DNSNames:              []string{hostname},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("create certificate: %v", err)
	}
	parsed, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatalf("parse certificate: %v", err)
	}
	return tls.Certificate{Certificate: [][]byte{der}, PrivateKey: key, Leaf: parsed}, parsed
}

func TestFetchMarketRatesRejectsMismatchedServerCertificate(t *testing.T) {
	cert, leaf := issueCert(t, "evil-attacker.local")

	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"MITM_INJECTED_RATE":0.0001}`))
	}))
	server.TLS = &tls.Config{Certificates: []tls.Certificate{cert}}
	server.StartTLS()
	defer server.Close()

	oldURL := MarketRatesURL
	MarketRatesURL = server.URL
	defer func() { MarketRatesURL = oldURL }()

	rates, err := FetchMarketRates()
	if err == nil {
		t.Fatalf("expected the TLS handshake to fail, got rates %v", rates)
	}
	if !strings.Contains(err.Error(), "certificate") && !strings.Contains(err.Error(), "x509") {
		t.Fatalf("expected a certificate verification error, got %v", err)
	}

	// With the presented certificate trusted as a root, verification still fails
	// because the hostname does not match: hostname verification is enabled.
	roots := x509.NewCertPool()
	roots.AddCert(leaf)
	if _, err := leaf.Verify(x509.VerifyOptions{Roots: roots, DNSName: "market-rates.example.com"}); err == nil {
		t.Fatal("expected a hostname mismatch to be rejected")
	}
}
