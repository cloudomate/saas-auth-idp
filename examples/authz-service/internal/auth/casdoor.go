package auth

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken       = errors.New("invalid token")
	ErrTokenExpired       = errors.New("token expired")
	ErrNoCertificate      = errors.New("no certificate configured")
	ErrInvalidCertificate = errors.New("invalid certificate")
)

// CasdoorClaims represents the JWT claims from Casdoor
type CasdoorClaims struct {
	jwt.RegisteredClaims
	Owner       string `json:"owner"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	Avatar      string `json:"avatar"`
	Tag         string `json:"tag"`
	Type        string `json:"type"`
	IsAdmin     bool   `json:"isAdmin"`
	IsGlobalAdmin bool `json:"isGlobalAdmin"`
}

// UserContext represents the authenticated user context
type UserContext struct {
	UserID        string
	Name          string
	Email         string
	Organization  string // Tenant
	IsAdmin       bool
	IsGlobalAdmin bool
}

// JWK represents a JSON Web Key
type JWK struct {
	Use string   `json:"use"`
	Kty string   `json:"kty"`
	Kid string   `json:"kid"`
	Alg string   `json:"alg"`
	N   string   `json:"n"`
	E   string   `json:"e"`
	X5c []string `json:"x5c"`
}

// JWKS represents a JSON Web Key Set
type JWKS struct {
	Keys []JWK `json:"keys"`
}

// CasdoorValidator validates Casdoor JWT tokens
type CasdoorValidator struct {
	endpoint     string
	organization string
	application  string
	publicKeys   map[string]*rsa.PublicKey
	httpClient   *http.Client
}

// NewCasdoorValidator creates a new Casdoor JWT validator
func NewCasdoorValidator(endpoint, organization, application string) (*CasdoorValidator, error) {
	v := &CasdoorValidator{
		endpoint:     strings.TrimSuffix(endpoint, "/"),
		organization: organization,
		application:  application,
		publicKeys:   make(map[string]*rsa.PublicKey),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}

	// Try to fetch the JWKS
	if err := v.fetchJWKS(); err != nil {
		return v, fmt.Errorf("failed to fetch JWKS: %w", err)
	}

	return v, nil
}

// ValidateToken validates a Casdoor JWT token and returns the user context
func (v *CasdoorValidator) ValidateToken(tokenString string) (*UserContext, error) {
	// Remove "Bearer " prefix if present
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")
	tokenString = strings.TrimPrefix(tokenString, "bearer ")

	if tokenString == "" {
		return nil, ErrInvalidToken
	}

	if len(v.publicKeys) == 0 {
		// Try to fetch JWKS again
		if err := v.fetchJWKS(); err != nil {
			return nil, fmt.Errorf("no public keys available: %w", err)
		}
	}

	// Parse and validate the token
	token, err := jwt.ParseWithClaims(tokenString, &CasdoorClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Get the key ID from token header
		kid, ok := token.Header["kid"].(string)
		if !ok {
			// If no kid, try the first key
			for _, key := range v.publicKeys {
				return key, nil
			}
			return nil, ErrNoCertificate
		}

		// Find the key by kid
		if key, exists := v.publicKeys[kid]; exists {
			return key, nil
		}

		return nil, fmt.Errorf("key not found for kid: %s", kid)
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	claims, ok := token.Claims.(*CasdoorClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return &UserContext{
		UserID:        claims.Name,
		Name:          claims.DisplayName,
		Email:         claims.Email,
		Organization:  claims.Owner,
		IsAdmin:       claims.IsAdmin,
		IsGlobalAdmin: claims.IsGlobalAdmin,
	}, nil
}

// fetchJWKS fetches the JSON Web Key Set from Casdoor
func (v *CasdoorValidator) fetchJWKS() error {
	url := fmt.Sprintf("%s/.well-known/jwks", v.endpoint)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to get JWKS: %s (status: %d)", string(body), resp.StatusCode)
	}

	var jwks JWKS
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return err
	}

	if len(jwks.Keys) == 0 {
		return ErrNoCertificate
	}

	// Parse all keys
	for _, jwk := range jwks.Keys {
		publicKey, err := jwkToPublicKey(jwk)
		if err != nil {
			continue // Skip invalid keys
		}
		v.publicKeys[jwk.Kid] = publicKey
	}

	if len(v.publicKeys) == 0 {
		return ErrNoCertificate
	}

	return nil
}

// jwkToPublicKey converts a JWK to an RSA public key
func jwkToPublicKey(jwk JWK) (*rsa.PublicKey, error) {
	// Try x5c certificate first
	if len(jwk.X5c) > 0 {
		certDER, err := base64.StdEncoding.DecodeString(jwk.X5c[0])
		if err != nil {
			return nil, err
		}
		cert, err := x509.ParseCertificate(certDER)
		if err != nil {
			return nil, err
		}
		if rsaKey, ok := cert.PublicKey.(*rsa.PublicKey); ok {
			return rsaKey, nil
		}
	}

	// Fall back to n and e
	if jwk.N != "" && jwk.E != "" {
		nBytes, err := base64.RawURLEncoding.DecodeString(jwk.N)
		if err != nil {
			return nil, err
		}
		eBytes, err := base64.RawURLEncoding.DecodeString(jwk.E)
		if err != nil {
			return nil, err
		}

		n := new(big.Int).SetBytes(nBytes)
		e := int(new(big.Int).SetBytes(eBytes).Int64())

		return &rsa.PublicKey{N: n, E: e}, nil
	}

	return nil, errors.New("unable to parse JWK")
}
