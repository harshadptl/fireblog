package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

// Signer produces tamper-proof cookie values with HMAC-SHA256.
type Signer struct{ key []byte }

// NewSigner returns a Signer for secret.
func NewSigner(secret string) Signer { return Signer{key: []byte(secret)} }

// Sign encodes v as JSON and appends a signature.
func (s Signer) Sign(v any) (string, error) {
	payload, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	enc := base64.RawURLEncoding.EncodeToString(payload)
	return enc + "." + s.mac(enc), nil
}

// Verify checks the signature and decodes the JSON payload into dst.
func (s Signer) Verify(signed string, dst any) error {
	enc, sig, ok := strings.Cut(signed, ".")
	if !ok || !hmac.Equal([]byte(sig), []byte(s.mac(enc))) {
		return errors.New("invalid signature")
	}
	payload, err := base64.RawURLEncoding.DecodeString(enc)
	if err != nil {
		return err
	}
	return json.Unmarshal(payload, dst)
}

func (s Signer) mac(data string) string {
	h := hmac.New(sha256.New, s.key)
	h.Write([]byte(data))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}

// OAuthState is stored in a short-lived signed cookie between the redirect
// to Google and the callback.
type OAuthState struct {
	State    string `json:"s"`
	Verifier string `json:"v"`
	Next     string `json:"n"`
	Issued   int64  `json:"t"`
}

// Expired reports whether the state is older than maxAge.
func (o OAuthState) Expired(maxAge time.Duration) bool {
	return time.Since(time.Unix(o.Issued, 0)) > maxAge
}

// SafeNext returns next if it's a same-site relative path, else "/".
// Prevents open redirects via ?next=https://evil.example.
func SafeNext(next string) string {
	if next == "" || !strings.HasPrefix(next, "/") || strings.HasPrefix(next, "//") ||
		strings.HasPrefix(next, "/\\") || strings.ContainsAny(next, "\r\n\t") {
		return "/"
	}
	return next
}
