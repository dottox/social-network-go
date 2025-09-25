package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type MockAuthenticator struct {
}

func NewMockAuthenticator() *MockAuthenticator {
	return &MockAuthenticator{}
}

const secret = "test"

var testClaims = jwt.MapClaims{
	"sub": uint32(24),                       // Subject
	"exp": time.Now().Add(time.Hour).Unix(), // Expiration time
	"iss": "test-iss",                       // Issuer
	"aud": "test-aud",                       // Audience
}

func (ma *MockAuthenticator) GenerateToken(claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, testClaims)
	tokenString, _ := token.SignedString([]byte(secret))

	return tokenString, nil
}

func (ma *MockAuthenticator) ValidateToken(token string) (*jwt.Token, error) {
	return jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
}
