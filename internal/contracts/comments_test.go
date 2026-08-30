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
