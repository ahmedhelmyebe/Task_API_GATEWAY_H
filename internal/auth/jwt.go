package auth // JWT issue/verify

import (
	"errors"       // For error values
	"time"         // For exp/iat

	"github.com/golang-jwt/jwt/v5" // JWT library
	"example.com/api-gateway/config" // JWT config
)

// Claims represents our custom JWT claims.
type Claims struct {
	Sub  string `json:"sub"`  // user id
	Role string `json:"role"` // user role (admin|user)
	jwt.RegisteredClaims       // iss, aud, iat, exp
}

// Sign builds a signed token string using HS256.
func Sign(c config.JWT, sub, role string) (string, error) {
	now := time.Now()
	claims := Claims{ // custom + registered
		Sub:  sub,
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    c.Issuer,
			Audience:  jwt.ClaimStrings{c.Audience},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(c.TTLMinutes) * time.Minute)),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims) // HS256 token
	return t.SignedString([]byte(c.Secret)) // HMAC sign
}

// Parse verifies signature and returns Claims.
func Parse(c config.JWT, token string) (*Claims, error) {
	parsed, err := jwt.ParseWithClaims(token, &Claims{}, func(t *jwt.Token) (any, error) {
		if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, errors.New("unexpected alg")
		}
		return []byte(c.Secret), nil
	})
	if err != nil {
		return nil, err
	}
	cl, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return nil, errors.New("invalid token")
	}
	return cl, nil
}