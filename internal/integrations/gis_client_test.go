package integrations

import (
	"net/http/httptest"
	"strings"
	"testing"
)

// A server presenting an untrusted, hostname-mismatched certificate must be
// rejected by the GIS client.
func TestFetchGeometryRejectsUntrustedCertificate(t *testing.T) {
	srv := httptest.NewTLSServer(nil)
	defer srv.Close()

	if _, err := FetchGeometry(srv.URL, "C-1001"); err == nil {
		t.Fatal("expected TLS verification error, got none")
	} else if !strings.Contains(err.Error(), "certificate") {
		t.Fatalf("expected certificate verification failure, got %v", err)
	}
}
