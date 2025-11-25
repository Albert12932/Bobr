package helpers

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

const LinkTokenTTL = 5 * time.Minute

func GenerateTokenRaw(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func HashToken(token string) []byte {
	sum := sha256.Sum256([]byte(token))
	return sum[:]
}

type JWTMaker struct {
	secret   []byte
	lifetime time.Duration
}

func NewJWTMaker(secret []byte, lifetime time.Duration) *JWTMaker {
	return &JWTMaker{secret: secret, lifetime: lifetime}
}

func (m *JWTMaker) Issue(userID, roleLevel int64) (token string, exp int64, err error) {
	exp = time.Now().Add(m.lifetime).Unix()

	claims := jwt.MapClaims{
		"sub":       userID, // кто
		"roleLevel": roleLevel,
		"exp":       exp, // срок
		"iat":       time.Now().Unix(),
	}

	j := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err = j.SignedString(m.secret)
	return
}

func (m *JWTMaker) Verify(token string) (jwt.MapClaims, error) {
	t, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		return m.secret, nil
	})
	if err != nil {
		return nil, err
	}
	if !t.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}

	return t.Claims.(jwt.MapClaims), nil
}

func NewRefreshToken() (string, error) {
	return GenerateTokenRaw(32) // 256 бит энтропии
}
