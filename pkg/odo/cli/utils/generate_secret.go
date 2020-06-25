package utils

import "crypto/rand"

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789$:#"

// GenerateSecureString creates a random secret of the desired length
func GenerateSecureString(length int) (string, error) {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	s := make([]byte, length)
	for i, v := range b {
		s[i] = charset[int(v)%len(charset)]
	}
	return string(s), nil
}
