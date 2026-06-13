package jwt

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Interface interface {
	GenerateAccessToken(input GenerateAccessTokenInput) (string, error)
	ParseAccessToken(tokenString string) (*Claims, error)
}

type jsonWebToken struct {
	SecretKey   string
	ExpiredTime time.Duration
}

type Claims struct {
	UserID        uuid.UUID `json:"user_id"`
	CooperativeID uuid.UUID `json:"cooperative_id"`
	SessionID     uuid.UUID `json:"session_id"`
	RoleCode      string    `json:"role_code"`
	jwt.RegisteredClaims
}

type GenerateAccessTokenInput struct {
	UserID        uuid.UUID
	CooperativeID uuid.UUID
	SessionID     uuid.UUID
	RoleCode      string
}

func Init() Interface {
	secretKey := os.Getenv("JWT_SECRET_KEY")
	expiredTime, err := strconv.Atoi(os.Getenv("JWT_EXP_TIME"))
	if err != nil {
		log.Fatalf("error init jwt %v", err)
	}

	return &jsonWebToken{
		SecretKey:   secretKey,
		ExpiredTime: time.Duration(expiredTime) * time.Hour,
	}
}

func (j *jsonWebToken) GenerateAccessToken(input GenerateAccessTokenInput) (string, error) {
	now := time.Now()

	claims := Claims{
		UserID:        input.UserID,
		CooperativeID: input.CooperativeID,
		SessionID:     input.SessionID,
		RoleCode:      input.RoleCode,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(j.ExpiredTime)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.SecretKey))
}

func (j *jsonWebToken) ParseAccessToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(j.SecretKey), nil
	})
	if err != nil || !token.Valid {
		return nil, err
	}
	return claims, nil
}
