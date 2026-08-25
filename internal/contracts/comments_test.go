package contracts_test

import (
	"testing"

	"github.com/cognition/sonar-remediation-demo/internal/contracts"
	"github.com/cognition/sonar-remediation-demo/internal/store"
)

func TestAddAndListComments(t *testing.T) {
	db, err := store.Open()
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer db.Close()

	if err := contracts.AddComment(db, "C-1002", "Ana Gomez", "Revisar el anexo"); err != nil {
		t.Fatalf("add comment: %v", err)
	}

	got, err := contracts.ListComments(db, "C-1002")
	if err != nil {
		t.Fatalf("list comments: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 comment, got %d", len(got))
	}
	if got[0].Body != "Revisar el anexo" {
		t.Fatalf("unexpected body %q", got[0].Body)
	}
}

// TestAddCommentDoesNotInject feeds AddComment the payloads that exploited the
// old fmt.Sprintf query: a second VALUES tuple and a stacked INSERT that copied
// the api_keys token into the thread. With bound parameters both must be stored
// verbatim as a single comment.
func TestAddCommentDoesNotInject(t *testing.T) {
	db, err := store.Open()
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer db.Close()

	payloads := []string{
		"x'), ('C-9001','INJECTED','pwned'); --",
		"x'); INSERT INTO comments (contract_id,author,body) SELECT 'C-9001','LEAK',token FROM api_keys; --",
	}

	for _, payload := range payloads {
		if err := contracts.AddComment(db, "C-9001", "Attacker", payload); err != nil {
			t.Fatalf("add comment with payload %q: %v", payload, err)
		}
	}

	got, err := contracts.ListComments(db, "C-9001")
	if err != nil {
		t.Fatalf("list comments: %v", err)
	}
	if len(got) != len(payloads) {
		t.Fatalf("expected %d comments, got %d: %+v", len(payloads), len(got), got)
	}
	for i, c := range got {
		if c.Author != "Attacker" {
			t.Fatalf("comment %d: injected author %q", i, c.Author)
		}
		if c.Body != payloads[i] {
			t.Fatalf("comment %d: body was interpreted, got %q want %q", i, c.Body, payloads[i])
		}
	}
}
