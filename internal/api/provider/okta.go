package provider

import (
	"context"
	"strings"

	"github.com/supabase/auth/internal/conf"
	"golang.org/x/oauth2"
)

type oktaProvider struct {
	*oauth2.Config
	APIUrl string
}

type oktaUser struct {
	Iss               string `json:"Iss"`
	Sub               string `json:"sub"`
	Aud               string `json:"aud"`
	Name              string `json:"name"`
	Nickname          string `json:"nickname"`
	PreferredUsername string `json:"preferred_username"`
	GivenName         string `json:"given_name"`
	MiddleName        string `json:"middle_name"`
	FamilyName        string `json:"family_name"`
	Email             string `json:"email"`
	EmailVerified     bool   `json:"email_verified"`
	Birthdate         string `json:"birthdate"`
	PhoneNumber       string `json:"phone_number"`
	Gender            string `json:"gender"`
	Picture           string `json:"picture"`
}

func NewOktaProvider(ext conf.OAuthProviderConfiguration, scopes string) (OAuthProvider, error) {
	if err := ext.ValidateOAuth(); err != nil {
		return nil, err
	}

	oauthScopes := []string{
		"openid",
		"email",
		"profile",
	}

	if scopes != "" {
		oauthScopes = append(oauthScopes, strings.Split(scopes, ",")...)
	}

	apiUrl := ext.URL

	return &oktaProvider{
		Config: &oauth2.Config{
			ClientID:     ext.ClientID[0],
			ClientSecret: ext.Secret,
			Endpoint: oauth2.Endpoint{
				AuthURL:  apiUrl + "/oauth2/v1/authorize",
				TokenURL: apiUrl + "/oauth2/v1/token",
			},
			RedirectURL: ext.RedirectURI,
			Scopes:      oauthScopes,
		},
		APIUrl: apiUrl,
	}, nil
}

func (g oktaProvider) GetOAuthToken(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error) {
	return g.Exchange(ctx, code, opts...)
}

func (g oktaProvider) RequiresPKCE() bool {
	return false
}

func (g oktaProvider) GetUserData(ctx context.Context, tok *oauth2.Token) (*UserProvidedData, error) {
	var u oktaUser
	if err := makeRequest(ctx, tok, g.Config, g.APIUrl+"/oauth2/v1/userinfo", &u); err != nil {
		return nil, err
	}

	data := &UserProvidedData{
		Metadata: &Claims{
			Issuer:            "okta",
			Subject:           u.Sub,
			Name:              u.Name,
			PreferredUsername: u.PreferredUsername,
			FullName:          u.Name,
			FamilyName:        u.FamilyName,
			GivenName:         u.GivenName,
			MiddleName:        u.MiddleName,
			NickName:          u.Nickname,
			Email:             strings.ToLower(u.Email),
			EmailVerified:     u.EmailVerified,
			Birthdate:         u.Birthdate,
			Phone:             u.PhoneNumber,
			Gender:            u.Gender,
			Picture:           u.Picture,
		},
		Emails: []Email{{
			Email:    u.Email,
			Primary:  true,
			Verified: true,
		}},
	}

	return data, nil
}
