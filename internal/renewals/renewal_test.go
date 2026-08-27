package renewals

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSignDownloadLink(t *testing.T) {
	link := SignDownloadLink("secret", "C-1001")
	if !strings.Contains(link, "id=C-1001") || !strings.Contains(link, "signature=") {
		t.Fatalf("unexpected download link %q", link)
	}
}

func TestSignDownloadLinkUsesHMAC(t *testing.T) {
	link := SignDownloadLink("secret", "C-1001")

	mac := hmac.New(sha256.New, []byte("secret"))
	mac.Write([]byte("C-1001"))
	want := "signature=" + hex.EncodeToString(mac.Sum(nil))
	if !strings.Contains(link, want) {
		t.Fatalf("expected HMAC-SHA256 signature %q in %q", want, link)
	}

	// A different secret or contract id must produce a different signature,
	// and the signature must not be a bare MD5 of secret+contractID.
	if link == SignDownloadLink("other-secret", "C-1001") {
		t.Fatal("signature must depend on the signing secret")
	}
	if link == SignDownloadLink("secret", "C-2002") {
		t.Fatal("signature must depend on the contract id")
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

func TestSigningSecret(t *testing.T) {
	if got := SigningSecret(); got == "" {
		t.Fatal("expected a signing secret")
	}
}
