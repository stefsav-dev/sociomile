package utils

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

var (
	AdminSecret = []byte(os.Getenv("JWT_ADMIN_SECRET"))
	AgentSecret = []byte(os.Getenv("JWT_AGENT_SECRET"))
	UserSecret  = []byte(os.Getenv("JWT_USER_SECRET"))
)

type TokenDetails struct {
	Token     string
	ExpiresAt time.Time
}

func GenerateToken(userID uint, role string) (*TokenDetails, error) {
	var secretKey []byte

	var expirationTime time.Time

	switch role {
	case "admin":
		secretKey = AdminSecret
		expirationTime = time.Now().Add(2 * time.Hour)
	case "agent":
		secretKey = AgentSecret
		expirationTime = time.Now().Add(10 * time.Hour)
	case "user":
		secretKey = UserSecret
		expirationTime = time.Now().Add(24 * time.Hour)
	default:
		secretKey = UserSecret
		expirationTime = time.Now().Add(24 * time.Hour)
	}

	claims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"exp":     expirationTime.Unix(),
		"iat":     time.Now().Unix(),
		"nbf":     time.Now().Unix(),
		"iss":     "sociomile-backend",
		"type":    "access",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secretKey)

	if err != nil {
		return nil, err
	}

	return &TokenDetails{
		Token:     tokenString,
		ExpiresAt: expirationTime,
	}, nil
}

func VerifyToken(tokenString string, role string) (*jwt.Token, jwt.MapClaims, error) {
	var secretKey []byte

	switch role {
	case "admin":
		secretKey = AdminSecret
	case "agent":
		secretKey = AgentSecret
	case "user":
		secretKey = UserSecret
	default:
		return nil, nil, errors.New("invalid role for token verification")
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return secretKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, nil, ErrExpiredToken
		}
		return nil, nil, ErrInvalidToken
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		tokenRole, ok := claims["role"].(string)
		if !ok || tokenRole != role {
			return nil, nil, errors.New("token role mismatch")
		}
		return token, claims, nil
	}

	return nil, nil, ErrInvalidToken
}

func GenerateRefreshToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}
