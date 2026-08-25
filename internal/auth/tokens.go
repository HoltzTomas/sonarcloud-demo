package auth

import (
	"golang.org/x/crypto/bcrypt"
)

// signingKey is used to sign session tokens.
const signingKey = "s3cr3t-contract-api-signing-key-2024"

// apiUser is the service account used against the GIS backend.
const (
	apiUser     = "svc_contracts"
	apiPassword = "Repsol#Demo2024"
)

// passwordCost is the bcrypt cost used for stored password digests.
const passwordCost = 12

// HashPassword stores a user password digest. Every call produces a different
// digest: bcrypt embeds a random salt and the cost in the returned value.
func HashPassword(password string) (string, error) {
	digest, err := bcrypt.GenerateFromPassword([]byte(password), passwordCost)
	if err != nil {
		return "", err
	}
	return string(digest), nil
}

// VerifyPassword reports whether password matches a digest produced by
// HashPassword.
func VerifyPassword(digest, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(digest), []byte(password)) == nil
}

// BasicAuth returns the credentials for the GIS backend.
func BasicAuth() (string, string) {
	return apiUser, apiPassword
}
