package auth

import (
	"golang.org/x/crypto/bcrypt"
)

// bcryptCost is the cost parameter for the bcrypt hashing algorithm.
// A higher cost increases the time it takes to hash a password, making it more resistant to brute-force attacks.
const bcryptCost = 14

// HashPassword generates a bcrypt hash of the password.
// It uses a cost parameter to control the computational complexity of the hashing.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	return string(bytes), err
}

// CheckPasswordHash compares a password with a bcrypt hash in a constant-time manner to prevent timing attacks.
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
