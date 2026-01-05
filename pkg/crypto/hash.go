// Package crypto provides cryptographic utilities.
package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
)

// HashSHA256 computes the SHA-256 hash of data and returns it as a hex string.
func HashSHA256(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// HashSHA256String computes the SHA-256 hash of a string and returns it as a hex string.
func HashSHA256String(s string) string {
	return HashSHA256([]byte(s))
}

// HashSHA256Reader computes the SHA-256 hash of data from a reader.
func HashSHA256Reader(r io.Reader) (string, error) {
	h := sha256.New()
	if _, err := io.Copy(h, r); err != nil {
		return "", fmt.Errorf("hash reader: %w", err)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// GenerateRandomBytes generates n random bytes.
func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return nil, fmt.Errorf("generate random bytes: %w", err)
	}
	return b, nil
}

// GenerateRandomHex generates a random hex string of length n*2.
func GenerateRandomHex(n int) (string, error) {
	b, err := GenerateRandomBytes(n)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// GenerateToken generates a secure random token suitable for URLs.
func GenerateToken() (string, error) {
	return GenerateRandomHex(32) // 64 characters
}

// GenerateVerificationCode generates a 6-digit verification code.
func GenerateVerificationCode() (string, error) {
	b := make([]byte, 3)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate verification code: %w", err)
	}
	// Convert to 6-digit number
	code := (int(b[0])<<16 | int(b[1])<<8 | int(b[2])) % 1000000
	return fmt.Sprintf("%06d", code), nil
}
