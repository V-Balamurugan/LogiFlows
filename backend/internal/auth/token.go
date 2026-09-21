package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken   = errors.New("invalid or malformed token")
	ErrExpiredToken   = errors.New("token has expired")
	ErrInvalidClaims  = errors.New("invalid token claims")
	ErrUnexpectedAlgo = errors.New("unexpected token signing method")
)

// Claims represents the JWT claims payload for LogiFlows.
type Claims struct {
	UserID          uuid.UUID `json:"user_id"`
	Email           string    `json:"email"`
	IsPlatformAdmin bool      `json:"is_platform_admin"`
	jwt.RegisteredClaims
}

// TokenService manages JWT token issuance and cryptographic verification.
type TokenService struct {
	secret []byte
	expiry time.Duration
	issuer string
}

// NewTokenService creates a new TokenService.
func NewTokenService(secret string, expiry time.Duration, issuer string) *TokenService {
	return &TokenService{
		secret: []byte(secret),
		expiry: expiry,
		issuer: issuer,
	}
}

// GenerateAccessToken generates a signed HMAC-SHA256 JWT access token.
func (s *TokenService) GenerateAccessToken(userID uuid.UUID, email string, isPlatformAdmin bool) (string, time.Time, error) {
	now := time.Now().UTC()
	expiresAt := now.Add(s.expiry)

	claims := Claims{
		UserID:          userID,
		Email:           email,
		IsPlatformAdmin: isPlatformAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			Issuer:    s.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			NotBefore: jwt.NewNumericDate(now),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(s.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, expiresAt, nil
}

// ValidateAccessToken parses, cryptographically verifies, and extracts claims from a JWT token.
func (s *TokenService) ValidateAccessToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		// Enforce strict HMAC signing algorithm to prevent algorithm confusion attacks
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("%w: %v", ErrUnexpectedAlgo, t.Header["alg"])
		}
		return s.secret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidClaims
	}

	return claims, nil
}
