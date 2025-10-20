package auth // Password hashing helpers

import "golang.org/x/crypto/bcrypt" // bcrypt implementation

// Hash returns bcrypt hash for a plaintext password.
func Hash(plain string) (string, error) { // cost default
	b, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	return string(b), err
}

// Verify compares plaintext with stored bcrypt hash.
func Verify(hash, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}