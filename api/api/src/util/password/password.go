package password

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// Hash returns the hash of the given password.
func Hash(password string) (string, error) {
	result, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", errors.New(fmt.Sprintf("failed to hash password: %s", err))
	}
	return string(result), nil
}

// Check returns true if the given password matches the hash, false otherwise.
func Check(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
