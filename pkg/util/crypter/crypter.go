package crypter

import (
	"github.com/segmentio/ksuid"
	"golang.org/x/crypto/bcrypt"
)

// HashPassword hashes the password using bcrypt
func HashPassword(password string) string {
	hashedPW, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hashedPW)
}

// CompareHashAndPassword matches hash with password. Returns true if hash and password match.
func CompareHashAndPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// UID returns unique string ID
func UID() string {
	return ksuid.New().String()
}
