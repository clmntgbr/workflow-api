package auth

import (
	"context"
	"errors"

	"go-api/internal/domain/port"
	domainuser "go-api/internal/domain/user"

	"github.com/golang-jwt/jwt/v5"
)

type ValidateTokenHandler struct {
	tokenKeys port.TokenKeyProvider
	userRepo  domainuser.UserWriteRepository
}

func NewValidateTokenHandler(
	tokenKeys port.TokenKeyProvider,
	userRepo domainuser.UserWriteRepository,
) *ValidateTokenHandler {
	return &ValidateTokenHandler{
		tokenKeys: tokenKeys,
		userRepo:  userRepo,
	}
}

func (uc *ValidateTokenHandler) Handle(ctx context.Context, input ValidateTokenInput) (*ValidateTokenOutput, error) {
	token, err := jwt.ParseWithClaims(
		input.Token,
		&JWTClaims{},
		uc.tokenKeys.GetKeyfunc(),
		jwt.WithIssuer(uc.tokenKeys.GetIssuer()),
		jwt.WithExpirationRequired(),
	)

	if err != nil {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	clerkID := claims.UserID
	if clerkID == "" {
		clerkID = claims.Subject
	}

	user, err := uc.userRepo.GetByClerkID(ctx, clerkID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	return &ValidateTokenOutput{
		User:   user,
		Claims: claims,
	}, nil
}
