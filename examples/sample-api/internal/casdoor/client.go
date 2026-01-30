package casdoor

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken     = errors.New("invalid token")
	ErrTokenExpired     = errors.New("token expired")
	ErrNoCertificate    = errors.New("no certificate configured")
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

// User represents a Casdoor user
type User struct {
	ID            string `json:"id"`
	Owner         string `json:"owner"`
	Name          string `json:"name"`
	DisplayName   string `json:"displayName"`
	Email         string `json:"email"`
	Phone         string `json:"phone"`
	Avatar        string `json:"avatar"`
	Type          string `json:"type"`
	IsAdmin       bool   `json:"isAdmin"`
	IsGlobalAdmin bool   `json:"isGlobalAdmin"`
	CreatedTime   string `json:"createdTime"`
}

// Client is a Casdoor API client
type Client struct {
	endpoint     string
	clientID     string
	clientSecret string
	organization string
	application  string
	publicKey    *rsa.PublicKey
	httpClient   *http.Client
}

// Config holds Casdoor client configuration
type Config struct {
	Endpoint     string
	ClientID     string
	ClientSecret string
	Organization string
	Application  string
	Certificate  string // PEM-encoded certificate
}

// NewClient creates a new Casdoor client
func NewClient(cfg Config) (*Client, error) {
	client := &Client{
		endpoint:     strings.TrimSuffix(cfg.Endpoint, "/"),
		clientID:     cfg.ClientID,
		clientSecret: cfg.ClientSecret,
		organization: cfg.Organization,
		application:  cfg.Application,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}

	// Parse certificate if provided
	if cfg.Certificate != "" {
		publicKey, err := parseCertificate(cfg.Certificate)
		if err != nil {
			return nil, fmt.Errorf("failed to parse certificate: %w", err)
		}
		client.publicKey = publicKey
	}

	return client, nil
}

// NewClientFromEnv creates a new Casdoor client from environment variables
func NewClientFromEnv() (*Client, error) {
	cfg := Config{
		Endpoint:     getEnv("CASDOOR_ENDPOINT", "http://localhost:8000"),
		ClientID:     os.Getenv("CASDOOR_CLIENT_ID"),
		ClientSecret: os.Getenv("CASDOOR_CLIENT_SECRET"),
		Organization: getEnv("CASDOOR_ORGANIZATION", "built-in"),
		Application:  getEnv("CASDOOR_APPLICATION", "app-built-in"),
		Certificate:  os.Getenv("CASDOOR_CERTIFICATE"),
	}

	return NewClient(cfg)
}

// ValidateToken validates a Casdoor JWT token and returns the claims
func (c *Client) ValidateToken(tokenString string) (*CasdoorClaims, error) {
	// Remove "Bearer " prefix if present
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")
	tokenString = strings.TrimPrefix(tokenString, "bearer ")

	if c.publicKey == nil {
		// If no certificate is configured, fetch it from Casdoor
		if err := c.fetchCertificate(); err != nil {
			return nil, fmt.Errorf("failed to fetch certificate: %w", err)
		}
	}

	// Parse and validate the token
	token, err := jwt.ParseWithClaims(tokenString, &CasdoorClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return c.publicKey, nil
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

	return claims, nil
}

// GetUser fetches user information from Casdoor API
func (c *Client) GetUser(name string) (*User, error) {
	url := fmt.Sprintf("%s/api/get-user?id=%s/%s", c.endpoint, c.organization, name)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Add client credentials for API authentication
	req.SetBasicAuth(c.clientID, c.clientSecret)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get user: %s", string(body))
	}

	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

// fetchCertificate fetches the certificate from Casdoor
func (c *Client) fetchCertificate() error {
	url := fmt.Sprintf("%s/api/get-application?id=%s/%s", c.endpoint, c.organization, c.application)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to get application: %s", string(body))
	}

	var result struct {
		Cert string `json:"cert"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	if result.Cert == "" {
		return ErrNoCertificate
	}

	publicKey, err := parseCertificate(result.Cert)
	if err != nil {
		return err
	}

	c.publicKey = publicKey
	return nil
}

// parseCertificate parses a PEM-encoded certificate and extracts the public key
func parseCertificate(certPEM string) (*rsa.PublicKey, error) {
	// Try to parse as certificate first
	block, _ := pem.Decode([]byte(certPEM))
	if block == nil {
		return nil, ErrInvalidCertificate
	}

	switch block.Type {
	case "CERTIFICATE":
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse certificate: %w", err)
		}
		if pubKey, ok := cert.PublicKey.(*rsa.PublicKey); ok {
			return pubKey, nil
		}
		return nil, errors.New("certificate does not contain RSA public key")

	case "PUBLIC KEY", "RSA PUBLIC KEY":
		pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse public key: %w", err)
		}
		if rsaPubKey, ok := pubKey.(*rsa.PublicKey); ok {
			return rsaPubKey, nil
		}
		return nil, errors.New("not an RSA public key")

	default:
		return nil, fmt.Errorf("unsupported PEM block type: %s", block.Type)
	}
}

// GetLoginURL returns the Casdoor login URL
func (c *Client) GetLoginURL(redirectURI, state string) string {
	return fmt.Sprintf("%s/login/oauth/authorize?client_id=%s&response_type=code&redirect_uri=%s&scope=read&state=%s",
		c.endpoint, c.clientID, redirectURI, state)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
