package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cognition/sonar-remediation-demo/internal/renewals"
	"github.com/cognition/sonar-remediation-demo/internal/store"
)

func TestRenewalHappyPathAndFingerprint(t *testing.T) {
	db, err := store.Open()
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer db.Close()

	market := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"energy":4.25,"freight":2.1}`))
	}))
	defer market.Close()
	oldURL := renewals.MarketRatesURL
	renewals.MarketRatesURL = market.URL
	defer func() { renewals.MarketRatesURL = oldURL }()

	req := httptest.NewRequest(http.MethodGet, "/contracts/renewal?id=C-1001", nil)
	res := httptest.NewRecorder()
	(&Server{DB: db}).Renewal(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.Code)
	}
	if res.Header().Get("ETag") == "" {
		t.Fatal("expected response fingerprint ETag")
	}
	if !strings.Contains(res.Body.String(), "Propuesta de renovacion") {
		t.Fatal("expected renewal proposal in response")
	}
}

func TestRenewalMissingID(t *testing.T) {
	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/contracts/renewal", nil)
	(&Server{}).Renewal(res, req)
	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", res.Code)
	}
}

func TestRenewalUnknownContract(t *testing.T) {
	db, err := store.Open()
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer db.Close()

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/contracts/renewal?id=C-9999", nil)
	(&Server{DB: db}).Renewal(res, req)
	if res.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", res.Code)
	}
}

func TestRenewalDegradesWhenMarketRatesAreUnavailable(t *testing.T) {
	db, err := store.Open()
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer db.Close()

	oldURL := renewals.MarketRatesURL
	renewals.MarketRatesURL = "http://127.0.0.1:1/unreachable"
	defer func() { renewals.MarketRatesURL = oldURL }()

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/contracts/renewal?id=C-1001", nil)
	(&Server{DB: db}).Renewal(res, req)
	if res.Code != http.StatusOK || !strings.Contains(res.Body.String(), "No se pudieron cargar") {
		t.Fatalf("expected fallback response, got %d: %s", res.Code, res.Body.String())
	}
}
