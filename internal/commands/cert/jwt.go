package certcmd

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// JWTInfo holds structured decoded JWT claims and metadata.
type JWTInfo struct {
	RawHeader  string         `json:"raw_header"`
	RawPayload string         `json:"raw_payload"`
	Header     map[string]any `json:"header"`
	Payload    map[string]any `json:"payload"`
	Algorithm  string         `json:"algorithm"`
	Type       string         `json:"type"`
	Subject    string         `json:"subject,omitempty"`
	Issuer     string         `json:"issuer,omitempty"`
	Audience   string         `json:"audience,omitempty"`
	IssuedAt   *time.Time     `json:"issued_at,omitempty"`
	ExpiresAt  *time.Time     `json:"expires_at,omitempty"`
	IsExpired  bool           `json:"is_expired"`
	TimeStatus string         `json:"time_status"`
	Signature  string         `json:"signature"`
}

// DecodeJWT parses and extracts claims from a JSON Web Token.
func DecodeJWT(token string) (*JWTInfo, error) {
	trimmed := strings.TrimSpace(token)
	parts := strings.Split(trimmed, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid JWT format: expected 3 dot-separated parts, got %d", len(parts))
	}

	headerBytes, err := decodeBase64Segment(parts[0])
	if err != nil {
		return nil, fmt.Errorf("decode JWT header: %w", err)
	}

	payloadBytes, err := decodeBase64Segment(parts[1])
	if err != nil {
		return nil, fmt.Errorf("decode JWT payload: %w", err)
	}

	var header map[string]any
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return nil, fmt.Errorf("unmarshal JWT header JSON: %w", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return nil, fmt.Errorf("unmarshal JWT payload JSON: %w", err)
	}

	info := &JWTInfo{
		RawHeader:  string(headerBytes),
		RawPayload: string(payloadBytes),
		Header:     header,
		Payload:    payload,
		Signature:  parts[2],
	}

	if alg, ok := header["alg"].(string); ok {
		info.Algorithm = alg
	}
	if typ, ok := header["typ"].(string); ok {
		info.Type = typ
	}
	if sub, ok := payload["sub"].(string); ok {
		info.Subject = sub
	}
	if iss, ok := payload["iss"].(string); ok {
		info.Issuer = iss
	}
	if aud, ok := payload["aud"].(string); ok {
		info.Audience = aud
	}

	now := time.Now()

	// Parse iat
	if iatVal, ok := payload["iat"].(float64); ok {
		t := time.Unix(int64(iatVal), 0)
		info.IssuedAt = &t
	}

	// Parse exp
	if expVal, ok := payload["exp"].(float64); ok {
		expTime := time.Unix(int64(expVal), 0)
		info.ExpiresAt = &expTime
		if now.After(expTime) {
			info.IsExpired = true
			diff := now.Sub(expTime).Round(time.Second)
			info.TimeStatus = fmt.Sprintf("Expired %s ago (%s)", diff, expTime.Format("2006-01-02 15:04:05 MST"))
		} else {
			info.IsExpired = false
			diff := expTime.Sub(now).Round(time.Second)
			info.TimeStatus = fmt.Sprintf("Valid for %s (until %s)", diff, expTime.Format("2006-01-02 15:04:05 MST"))
		}
	} else {
		info.TimeStatus = "No expiration (exp) claim"
	}

	return info, nil
}

func decodeBase64Segment(seg string) ([]byte, error) {
	// Add padding if required
	switch len(seg) % 4 {
	case 2:
		seg += "=="
	case 3:
		seg += "="
	}
	// Try URL encoding first
	data, err := base64.URLEncoding.DecodeString(seg)
	if err == nil {
		return data, nil
	}
	// Fallback to standard encoding
	return base64.StdEncoding.DecodeString(seg)
}
