package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"google.golang.org/api/idtoken"
)

type GoogleClaims struct {
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Sub           string `json:"sub"`
}

type GoogleOAuth struct {
	clientId string
}

func (g *GoogleOAuth) Verify(ctx context.Context, credential string) (*OAuthInfo, error) {
	payload, err := idtoken.Validate(ctx, credential, g.clientId)
	if err != nil {
		return nil, fmt.Errorf("invalid google token: %w", err)
	}

	claims := &GoogleClaims{}
	if err := decodeClaims(payload, claims); err != nil {
		return nil, fmt.Errorf("failed to extract claims: %w", err)
	}

	if !claims.EmailVerified {
		return nil, fmt.Errorf("google email not verified")
	}

	return &OAuthInfo{
		Email:         claims.Email,
		EmailVerified: claims.EmailVerified,
		Name:          claims.Name,
		GivenName:     claims.GivenName,
		FamilyName:    claims.FamilyName,
		ProviderId:    claims.Sub,
	}, nil
}

func decodeClaims(payload *idtoken.Payload, target any) error {
	b, err := json.Marshal(payload.Claims)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, target)
}

func NewGoogleOAuth(clientId string) Provider {
	return &GoogleOAuth{clientId: clientId}
}
