package clerk

import (
	"fmt"
	"go-api/internal/infrastructure/config"
	"time"

	"github.com/MicahParks/keyfunc/v2"
	"github.com/golang-jwt/jwt/v5"
)

type JWKSProvider struct {
	jwks   *keyfunc.JWKS
	issuer string
}

func NewJWKSProvider(cfg *config.Config) (*JWKSProvider, error) {
	jwksURL := fmt.Sprintf("%s/.well-known/jwks.json", cfg.ClerkFrontendAPI)

	jwks, err := keyfunc.Get(jwksURL, keyfunc.Options{
		RefreshInterval: time.Hour,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch JWKS: %w", err)
	}

	return &JWKSProvider{
		jwks:   jwks,
		issuer: cfg.ClerkFrontendAPI,
	}, nil
}

func (p *JWKSProvider) GetKeyfunc() jwt.Keyfunc {
	return p.jwks.Keyfunc
}

func (p *JWKSProvider) GetIssuer() string {
	return p.issuer
}

func (p *JWKSProvider) Cleanup() {
	p.jwks.EndBackground()
}
