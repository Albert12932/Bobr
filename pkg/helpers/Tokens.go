package helpers

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const LinkTokenTTL = 5 * time.Minute

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

type JWTMaker struct {
	secret   []byte
	lifetime time.Duration
}

func NewJWTMaker(secret []byte, lifetime time.Duration) *JWTMaker {
	return &JWTMaker{secret: secret, lifetime: lifetime}
}

func (m *JWTMaker) Issue(userID int64, bookID int, firstName, surname string) (token string, exp time.Time, err error) {
	exp = time.Now().Add(m.lifetime)

	claims := jwt.MapClaims{
		"sub":       userID, // кто
		"book_id":   bookID,
		"firstName": firstName,
		"surname":   surname,
		"exp":       exp.Unix(), // срок
		"iat":       time.Now().Unix(),
	}

	j := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err = j.SignedString(m.secret)
	return
}
func (m *JWTMaker) Verify(ctx context.Context, token string) (jwt.MapClaims, error) {
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
