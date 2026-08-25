package web

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/cognition/sonar-remediation-demo/internal/store"
)

func newTestServer(t *testing.T) *Server {
	t.Helper()
	db, err := store.Open()
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return &Server{DB: db}
}

func TestInitial(t *testing.T) {
	cases := map[string]string{
		"":            "?",
		"ana gomez":   "A",
		"Luis Martin": "L",
	}
	for author, want := range cases {
		if got := initial(author); got != want {
			t.Errorf("initial(%q) = %q, want %q", author, got, want)
		}
	}
}

func TestAvatarColorIsFromPalette(t *testing.T) {
	for i := 0; i < 20; i++ {
		got := avatarColor()
		found := false
		for _, c := range avatarPalette {
			if got == c {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("avatarColor() = %q, not in palette", got)
		}
	}
}

func TestCommentsGetRendersThread(t *testing.T) {
	s := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/contracts/comments?contract=C-1001", nil)
	rec := httptest.NewRecorder()
	s.Comments(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	for _, want := range []string{"Seguimiento de C-1001", "Ana Gomez", "Revisada la clausula", "Luis Martin"} {
		if !strings.Contains(body, want) {
			t.Errorf("body does not contain %q", want)
		}
	}
}

func TestCommentsGetEmptyThread(t *testing.T) {
	s := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/contracts/comments?contract=C-1003", nil)
	rec := httptest.NewRecorder()
	s.Comments(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Todavia no hay comentarios") {
		t.Error("expected the empty-thread message")
	}
}

func TestCommentsMissingContract(t *testing.T) {
	s := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/contracts/comments?contract=%20", nil)
	rec := httptest.NewRecorder()
	s.Comments(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestCommentsPostStoresAndRedirects(t *testing.T) {
	s := newTestServer(t)

	form := url.Values{
		"contract": {"C-1004"},
		"author":   {"Marta Ruiz"},
		"body":     {"Falta el anexo de precios"},
	}
	post := httptest.NewRequest(http.MethodPost, "/contracts/comments", strings.NewReader(form.Encode()))
	post.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	s.Comments(rec, post)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want 303", rec.Code)
	}
	if loc := rec.Header().Get("Location"); loc != "/contracts/comments?contract=C-1004" {
		t.Fatalf("redirect to %q", loc)
	}

	get := httptest.NewRequest(http.MethodGet, "/contracts/comments?contract=C-1004", nil)
	rec = httptest.NewRecorder()
	s.Comments(rec, get)
	body := rec.Body.String()
	if !strings.Contains(body, "Marta Ruiz") || !strings.Contains(body, "Falta el anexo de precios") {
		t.Fatalf("stored comment not rendered: %s", body)
	}
}

func TestRoutesExposesCommentsPage(t *testing.T) {
	s := newTestServer(t)
	mux := s.Routes()

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/contracts/comments?contract=C-1001", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Seguimiento de C-1001") {
		t.Error("comments page not served through the router")
	}
}
