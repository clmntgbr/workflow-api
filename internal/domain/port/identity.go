package port

import (
	"context"

	"github.com/golang-jwt/jwt/v5"
)

type ClerkUser struct {
	ID        string
	FirstName string
	LastName  string
	Email     string
	Banned    bool
}

type ClerkUserFetcher interface {
	Get(ctx context.Context, clerkID string) (ClerkUser, error)
}

type TokenKeyProvider interface {
	GetKeyfunc() jwt.Keyfunc
	GetIssuer() string
}
