package oauth

import "context"

type OAuthInfo struct {
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	ProviderId    string `json:"provider_id"`
}

type Provider interface {
	Verify(ctx context.Context, credential string) (*OAuthInfo, error)
}
