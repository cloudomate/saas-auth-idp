package auth

import (
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

type JWTValidator struct {
	secret []byte
}

func NewJWTValidator(secret []byte) *JWTValidator {
	return &JWTValidator{secret: secret}
}

type JWTClaims struct {
	jwt.RegisteredClaims
	Email           string `json:"email"`
	Name            string `json:"name"`
	TenantID        string `json:"tenant_id"`
	IsPlatformAdmin bool   `json:"is_platform_admin"`
	IsTenantAdmin   bool   `json:"is_tenant_admin"`
}

func (v *JWTValidator) Validate(tokenString string) (*Identity, error) {
	if len(v.secret) == 0 {
		return nil, errors.New("jwt secret not configured")
	}

	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return v.secret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	return &Identity{
		UserID:          claims.Subject,
		Email:           claims.Email,
		TenantID:        claims.TenantID,
		IsPlatformAdmin: claims.IsPlatformAdmin,
	}, nil
}
