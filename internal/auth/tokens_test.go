package auth

import "testing"

func TestHashPasswordIsSaltedAndVerifiable(t *testing.T) {
	first, err := HashPassword("Repsol#Demo2024")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	second, err := HashPassword("Repsol#Demo2024")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	if first == second {
		t.Fatal("digests are deterministic, so they are not salted")
	}
	if !VerifyPassword(first, "Repsol#Demo2024") {
		t.Fatal("digest does not verify against its own password")
	}
	if VerifyPassword(first, "wrong-password") {
		t.Fatal("digest verifies against a wrong password")
	}
}
