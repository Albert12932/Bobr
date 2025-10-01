package helpers

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"time"
)

const LinkTokenTTL = 5 * time.Minute

// helpers

func GenerateTokenRaw(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	// URL-safe без '=' (короче и удобнее в ссылках)
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func HashToken(token string) []byte {
	sum := sha256.Sum256([]byte(token))
	return sum[:]
}
