package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/newfeed/community-news/services/auth-service/internal/auth/domain"
)

type JWTSigner struct {
	secret   []byte
	ttl      time.Duration
	issuer   string
	audience string
}

func NewJWTSigner(secret string, ttl time.Duration) *JWTSigner {
	return &JWTSigner{secret: []byte(secret), ttl: ttl, issuer: "newfeed-auth-service", audience: "newfeed-api"}
}

func (s *JWTSigner) SignAccessToken(user domain.User) (string, time.Time, error) {
	if len(s.secret) < 32 {
		return "", time.Time{}, errors.New("jwt secret must be at least 32 bytes")
	}
	now := time.Now().UTC()
	expiresAt := time.Now().UTC().Add(s.ttl)
	header := map[string]string{"alg": "HS256", "typ": "JWT"}
	claims := map[string]any{
		"sub":   user.ID,
		"email": user.Email,
		"role":  string(user.Role),
		"iss":   s.issuer,
		"aud":   s.audience,
		"exp":   expiresAt.Unix(),
		"iat":   now.Unix(),
		"nbf":   now.Unix(),
	}
	headerPart, err := encodeJSON(header)
	if err != nil {
		return "", time.Time{}, err
	}
	claimPart, err := encodeJSON(claims)
	if err != nil {
		return "", time.Time{}, err
	}
	payload := headerPart + "." + claimPart
	return payload + "." + s.sign(payload), expiresAt, nil
}

func (s *JWTSigner) ParseAccessToken(raw string) (*domain.Claims, error) {
	parts := strings.Split(raw, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid token")
	}
	header, err := decodeMap(parts[0])
	if err != nil {
		return nil, err
	}
	if stringClaim(header["alg"]) != "HS256" || stringClaim(header["typ"]) != "JWT" {
		return nil, errors.New("invalid token header")
	}
	payload := parts[0] + "." + parts[1]
	expected, err := base64.RawURLEncoding.DecodeString(s.sign(payload))
	if err != nil {
		return nil, err
	}
	actual, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, err
	}
	if !hmac.Equal(actual, expected) {
		return nil, errors.New("invalid token signature")
	}
	claims, err := decodeMap(parts[1])
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC().Unix()
	exp, err := numericClaim(claims["exp"])
	if err != nil || now >= exp {
		return nil, errors.New("token expired")
	}
	nbf, err := numericClaim(claims["nbf"])
	if err != nil || now < nbf {
		return nil, errors.New("token not active")
	}
	if stringClaim(claims["iss"]) != s.issuer || stringClaim(claims["aud"]) != s.audience {
		return nil, errors.New("invalid token issuer or audience")
	}
	return &domain.Claims{
		UserID: stringClaim(claims["sub"]),
		Email:  stringClaim(claims["email"]),
		Role:   stringClaim(claims["role"]),
	}, nil
}

func decodeMap(part string) (map[string]any, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(part)
	if err != nil {
		return nil, err
	}
	var out map[string]any
	if err := json.Unmarshal(decoded, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *JWTSigner) sign(payload string) string {
	mac := hmac.New(sha256.New, s.secret)
	mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func encodeJSON(value any) (string, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(data), nil
}

func stringClaim(value any) string {
	if text, ok := value.(string); ok {
		return text
	}
	return ""
}

func numericClaim(value any) (int64, error) {
	switch typed := value.(type) {
	case float64:
		return int64(typed), nil
	case string:
		return strconv.ParseInt(typed, 10, 64)
	default:
		return 0, errors.New("invalid numeric claim")
	}
}
