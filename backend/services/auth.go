package services

import (
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"

	"golang.org/x/crypto/bcrypt"
)

var tokenKey []byte

// SetTokenKey sets the key used to sign participant tokens. Must be called once at startup.
func SetTokenKey(key []byte) { tokenKey = key }

// participantToken is a stateless credential the browser keeps instead of the
// password. It is bound to the stored bcrypt hash, so it stops working if the
// participant's password ever changes, and checking it costs one HMAC instead
// of a bcrypt comparison.
func participantToken(sessionID, name, hash string) string {
	if hash == "" {
		return ""
	}
	mac := hmac.New(sha256.New, tokenKey)
	mac.Write([]byte(sessionID))
	mac.Write([]byte{0})
	mac.Write([]byte(name))
	mac.Write([]byte{0})
	mac.Write([]byte(hash))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

// checkCredential verifies a token or password against a stored hash.
// A missing hash means the resource has no password and anyone may act on it.
func checkCredential(sessionID, name string, hash sql.NullString, password, token string) error {
	if !hash.Valid || hash.String == "" {
		return nil
	}
	if token != "" && hmac.Equal([]byte(token), []byte(participantToken(sessionID, name, hash.String))) {
		return nil
	}
	if password == "" {
		return ErrPasswordRequired
	}
	if bcrypt.CompareHashAndPassword([]byte(hash.String), []byte(password)) != nil {
		return ErrInvalidPassword
	}
	return nil
}

func hashPassword(password string) (sql.NullString, error) {
	if password == "" {
		return sql.NullString{}, nil
	}
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return sql.NullString{}, err
	}
	return sql.NullString{String: string(b), Valid: true}, nil
}
