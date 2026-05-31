package security

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"strings"
)

type PasswordHasher struct{}

func NewPasswordHasher() PasswordHasher {
	return PasswordHasher{}
}

func (PasswordHasher) Hash(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	hash := passwordDigest(salt, password)
	return hex.EncodeToString(salt) + ":" + hex.EncodeToString(hash), nil
}

func (PasswordHasher) Compare(encoded, password string) bool {
	parts := strings.Split(encoded, ":")
	if len(parts) != 2 {
		return false
	}
	salt, err := hex.DecodeString(parts[0])
	if err != nil {
		return false
	}
	expected, err := hex.DecodeString(parts[1])
	if err != nil {
		return false
	}
	actual := passwordDigest(salt, password)
	return subtle.ConstantTimeCompare(expected, actual) == 1
}

func passwordDigest(salt []byte, password string) []byte {
	sum := sha256.Sum256(append(salt, []byte(password)...))
	for i := 0; i < 20000; i++ {
		sum = sha256.Sum256(sum[:])
	}
	return sum[:]
}
