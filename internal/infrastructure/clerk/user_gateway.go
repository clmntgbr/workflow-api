package clerk

import (
	"context"
	"errors"

	"go-api/internal/domain/port"

	"github.com/clerk/clerk-sdk-go/v2"
	clerkuser "github.com/clerk/clerk-sdk-go/v2/user"
)

type UserGateway struct{}

func NewUserGateway(secretKey string) *UserGateway {
	clerk.SetKey(secretKey)
	return &UserGateway{}
}

func (g *UserGateway) Get(ctx context.Context, clerkID string) (port.ClerkUser, error) {
	clerkUser, err := clerkuser.Get(ctx, clerkID)
	if err != nil {
		return port.ClerkUser{}, errors.New("failed to get user")
	}

	firstName := ""
	if clerkUser.FirstName != nil {
		firstName = *clerkUser.FirstName
	}

	lastName := ""
	if clerkUser.LastName != nil {
		lastName = *clerkUser.LastName
	}

	email := ""
	for _, address := range clerkUser.EmailAddresses {
		if clerkUser.PrimaryEmailAddressID != nil && address.ID == *clerkUser.PrimaryEmailAddressID {
			email = address.EmailAddress
			break
		}
	}
	if email == "" && len(clerkUser.EmailAddresses) > 0 {
		email = clerkUser.EmailAddresses[0].EmailAddress
	}

	return port.ClerkUser{
		ID:        clerkUser.ID,
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
		Banned:    clerkUser.Banned,
	}, nil
}
