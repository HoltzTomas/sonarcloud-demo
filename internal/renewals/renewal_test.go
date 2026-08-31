package renewals

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func signatureOf(t *testing.T, link string) string {
	t.Helper()
	_, signature, ok := strings.Cut(link, "signature=")
	if !ok {
		t.Fatalf("no signature in download link %q", link)
	}
	return signature
}

func TestSignDownloadLink(t *testing.T) {
	link := SignDownloadLink("secret", "C-1001")
	if !strings.Contains(link, "id=C-1001") || !strings.Contains(link, "signature=") {
		t.Fatalf("unexpected download link %q", link)
	}

	mac := hmac.New(sha256.New, []byte("secret"))
	mac.Write([]byte("C-1001"))
	want := fmt.Sprintf("%x", mac.Sum(nil))
	if got := signatureOf(t, link); got != want {
		t.Fatalf("signature is not HMAC-SHA256 of the contract id: got %q, want %q", got, want)
	}
}

func TestSignDownloadLinkDependsOnSecret(t *testing.T) {
	if signatureOf(t, SignDownloadLink("secret", "C-1001")) == signatureOf(t, SignDownloadLink("other-secret", "C-1001")) {
		t.Fatal("the signature must change when the signing secret changes")
	}
}

func TestSignDownloadLinkIsNotAmbiguousAcrossContracts(t *testing.T) {
	// A plain hash of secret+contractID lets "C-100" + "1" and "C-1001" + ""
	// collide; keying the MAC over the contract id alone must not.
	if signatureOf(t, SignDownloadLink("secretC-100", "1")) == signatureOf(t, SignDownloadLink("secret", "C-1001")) {
		t.Fatal("signatures for different contracts must not collide")
	}
}

func TestFetchMarketRates(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/error" {
			http.Error(w, "unavailable", http.StatusBadGateway)
			return
		}
		_, _ = w.Write([]byte(`{"energy":4.25}`))
	}))
	defer server.Close()

	oldURL := MarketRatesURL
	defer func() { MarketRatesURL = oldURL }()

	MarketRatesURL = server.URL
	rates, err := FetchMarketRates()
	if err != nil || rates["energy"] != 4.25 {
		t.Fatalf("unexpected rates result: %v, %v", rates, err)
	}

	MarketRatesURL = server.URL + "/error"
	if _, err := FetchMarketRates(); err == nil {
		t.Fatal("expected an error for a non-200 response")
	}

	MarketRatesURL = "http://127.0.0.1:1/unreachable"
	if _, err := FetchMarketRates(); err == nil {
		t.Fatal("expected an error for an unavailable service")
	}
}

func TestFetchMarketRatesRejectsUntrustedCertificate(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"energy":6.66}`))
	}))
	defer server.Close()

	oldURL := MarketRatesURL
	defer func() { MarketRatesURL = oldURL }()
	MarketRatesURL = server.URL

	rates, err := FetchMarketRates()
	if err == nil {
		t.Fatalf("expected the untrusted certificate to be rejected, got rates %v", rates)
	}
	if !strings.Contains(err.Error(), "certificate") {
		t.Fatalf("expected a certificate validation error, got %v", err)
	}
}

func TestSigningSecret(t *testing.T) {
	if got := SigningSecret(); got == "" {
		t.Fatal("expected a signing secret")
	}
}
