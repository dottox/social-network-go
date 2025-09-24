package auth

import (
	"github.com/golang-jwt/jwt/v5"
)

type JWTAuthenticator struct {
	Secret string
	Aud    string
	Iss    string
}

func NewJWTAuthenticator(secret, audience, issuer string) *JWTAuthenticator {
	return &JWTAuthenticator{
		Secret: secret,
		Aud:    audience,
		Iss:    issuer,
	}
}

func (auth *JWTAuthenticator) GenerateToken(claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(auth.Secret))
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

func (auth *JWTAuthenticator) ValidateToken(token string) (*jwt.Token, error) {
	return jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(auth.Secret), nil
	},
		jwt.WithExpirationRequired(),
		jwt.WithAudience(auth.Aud),
		jwt.WithIssuer(auth.Iss),
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}),
	)
}
