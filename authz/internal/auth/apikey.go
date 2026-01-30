package auth

import (
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

const KeyPrefix = "sk"

var (
	ErrInvalidFormat  = errors.New("invalid API key format")
	ErrHashMismatch   = errors.New("API key hash does not match")
	ErrKeyNotFound    = errors.New("API key not found")
	ErrKeyRevoked     = errors.New("API key has been revoked")
	ErrKeyExpired     = errors.New("API key has expired")
)

type APIKeyValidator struct {
	db     *sql.DB
	secret []byte
}

func NewAPIKeyValidator(databaseURL string, secret []byte) (*APIKeyValidator, error) {
	if databaseURL == "" {
		return nil, errors.New("database URL required for API key validation")
	}

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &APIKeyValidator{
		db:     db,
		secret: secret,
	}, nil
}

func (v *APIKeyValidator) Close() error {
	if v.db != nil {
		return v.db.Close()
	}
	return nil
}

func (v *APIKeyValidator) Validate(token string) (*Identity, error) {
	// Parse key to get key ID
	keyID, err := v.parseKey(token)
	if err != nil {
		return nil, err
	}

	// Look up key from database
	identity, keyHash, err := v.lookupKey(keyID)
	if err != nil {
		return nil, err
	}

	// Validate hash
	if err := v.validateHash(token, keyHash); err != nil {
		return nil, err
	}

	identity.KeyID = keyID
	return identity, nil
}

func (v *APIKeyValidator) parseKey(token string) (string, error) {
	// Format: sk-<key_id>-<secret>
	parts := strings.SplitN(token, "-", 3)
	if len(parts) < 3 {
		return "", ErrInvalidFormat
	}

	prefix, keyID := parts[0], parts[1]

	if prefix != KeyPrefix {
		return "", ErrInvalidFormat
	}

	if len(keyID) != 8 && len(keyID) != 16 {
		return "", ErrInvalidFormat
	}

	if _, err := hex.DecodeString(keyID); err != nil {
		return "", ErrInvalidFormat
	}

	return keyID, nil
}

func (v *APIKeyValidator) validateHash(token string, storedHash string) error {
	h := sha256.Sum256([]byte(token))
	providedHash := hex.EncodeToString(h[:])

	if subtle.ConstantTimeCompare([]byte(providedHash), []byte(storedHash)) != 1 {
		return ErrHashMismatch
	}

	return nil
}

func (v *APIKeyValidator) lookupKey(keyID string) (*Identity, string, error) {
	query := `
		SELECT
			ak.user_id,
			ak.tenant_id,
			ak.workspace_id,
			ak.role,
			ak.key_hash,
			ak.revoked_at,
			ak.expires_at,
			u.email,
			u.is_platform_admin
		FROM api_keys ak
		LEFT JOIN users u ON ak.user_id = u.id
		WHERE ak.key_id = $1
	`

	var userID, tenantID string
	var workspaceID sql.NullString
	var role string
	var keyHash sql.NullString
	var revokedAt sql.NullTime
	var expiresAt sql.NullTime
	var email sql.NullString
	var isPlatformAdmin bool

	err := v.db.QueryRow(query, keyID).Scan(
		&userID,
		&tenantID,
		&workspaceID,
		&role,
		&keyHash,
		&revokedAt,
		&expiresAt,
		&email,
		&isPlatformAdmin,
	)

	if err == sql.ErrNoRows {
		return nil, "", ErrKeyNotFound
	}
	if err != nil {
		return nil, "", err
	}

	if revokedAt.Valid {
		return nil, "", ErrKeyRevoked
	}

	if expiresAt.Valid && expiresAt.Time.Before(time.Now()) {
		return nil, "", ErrKeyExpired
	}

	identity := &Identity{
		UserID:          userID,
		TenantID:        tenantID,
		Role:            role,
		IsPlatformAdmin: isPlatformAdmin,
	}

	if workspaceID.Valid {
		identity.WorkspaceID = workspaceID.String
	}

	if email.Valid {
		identity.Email = email.String
	}

	hash := ""
	if keyHash.Valid {
		hash = keyHash.String
	}

	return identity, hash, nil
}
