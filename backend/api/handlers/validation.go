package handlers

import "fmt"

const minPasswordLength = 8

// validatePassword checks that a password meets minimum security requirements.
func validatePassword(password string) error {
	if len(password) < minPasswordLength {
		return fmt.Errorf("password must be at least %d characters", minPasswordLength)
	}
	return nil
}

// validateUsername checks that a username is valid.
func validateUsername(username string) error {
	if len(username) < 3 {
		return fmt.Errorf("username must be at least 3 characters")
	}
	if len(username) > 64 {
		return fmt.Errorf("username must be at most 64 characters")
	}
	return nil
}
