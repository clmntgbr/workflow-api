package auth

import (
	domainuser "go-api/internal/domain/user"

	"github.com/golang-jwt/jwt/v5"
)

type JWTClaims struct {
	Email  string `json:"email"`
	UserID string `json:"sub"`
	jwt.RegisteredClaims
}

type ValidateTokenInput struct {
	Token string
}

type ValidateTokenOutput struct {
	User   *domainuser.User
	Claims *JWTClaims
}
