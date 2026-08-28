package web_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/cognition/sonar-remediation-demo/internal/store"
	"github.com/cognition/sonar-remediation-demo/internal/web"
)

func TestCommentsGetAndPost(t *testing.T) {
	db, err := store.Open()
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer db.Close()

	server := &web.Server{DB: db}

	getRequest := httptest.NewRequest(http.MethodGet, "/contracts/comments?contract=C-1001", nil)
	getResponse := httptest.NewRecorder()
	server.Comments(getResponse, getRequest)

	if getResponse.Code != http.StatusOK {
		t.Fatalf("GET status = %d, want %d", getResponse.Code, http.StatusOK)
	}
	body := getResponse.Body.String()
	if !strings.Contains(body, "Revisada la clausula de revision de precios.") {
		t.Fatalf("GET body does not contain the seeded comment: %s", body)
	}

	form := url.Values{
		"contract": {"C-1001"},
		"author":   {"Carlos Ruiz"},
		"body":     {"Llamar al cliente el viernes."},
	}
	postRequest := httptest.NewRequest(
		http.MethodPost,
		"/contracts/comments",
		strings.NewReader(form.Encode()),
	)
	postRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	postResponse := httptest.NewRecorder()
	server.Comments(postResponse, postRequest)

	if postResponse.Code != http.StatusSeeOther {
		t.Fatalf("POST status = %d, want %d", postResponse.Code, http.StatusSeeOther)
	}
	if location := postResponse.Header().Get("Location"); location != "/contracts/comments?contract=C-1001" {
		t.Fatalf("POST Location = %q, want %q", location, "/contracts/comments?contract=C-1001")
	}

	missingContractRequest := httptest.NewRequest(http.MethodGet, "/contracts/comments", nil)
	missingContractResponse := httptest.NewRecorder()
	server.Comments(missingContractResponse, missingContractRequest)
	if missingContractResponse.Code != http.StatusBadRequest {
		t.Fatalf("missing contract status = %d, want %d", missingContractResponse.Code, http.StatusBadRequest)
	}
}
