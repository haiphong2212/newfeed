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
	secret []byte
	ttl    time.Duration
}

func NewJWTSigner(secret string, ttl time.Duration) *JWTSigner {
	return &JWTSigner{secret: []byte(secret), ttl: ttl}
}

func (s *JWTSigner) SignAccessToken(user domain.User) (string, time.Time, error) {
	expiresAt := time.Now().UTC().Add(s.ttl)
	header := map[string]string{"alg": "HS256", "typ": "JWT"}
	claims := map[string]any{
		"sub":   user.ID,
		"email": user.Email,
		"role":  string(user.Role),
		"exp":   expiresAt.Unix(),
		"iat":   time.Now().UTC().Unix(),
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
	payload := parts[0] + "." + parts[1]
	if !hmac.Equal([]byte(parts[2]), []byte(s.sign(payload))) {
		return nil, errors.New("invalid token signature")
	}
	decoded, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}
	var claims map[string]any
	if err := json.Unmarshal(decoded, &claims); err != nil {
		return nil, err
	}
	exp, err := numericClaim(claims["exp"])
	if err != nil || time.Now().UTC().Unix() >= exp {
		return nil, errors.New("token expired")
	}
	return &domain.Claims{
		UserID: stringClaim(claims["sub"]),
		Email:  stringClaim(claims["email"]),
		Role:   stringClaim(claims["role"]),
	}, nil
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
