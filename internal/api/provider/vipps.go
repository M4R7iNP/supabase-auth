package provider

import (
	"context"
	"strings"

	"github.com/supabase/auth/internal/conf"
	"golang.org/x/oauth2"
)

const (
	defaultVippsAPIBase = "api.vipps.no"
)

type vippsProvider struct {
	*oauth2.Config
	APIPath string
}

type vippsUser struct {
	FamilyName string `json:"family_name"`
	GivenName  string `json:"given_name"`
	Name       string `json:"name"`
	Email      string `json:"email"`
	Sid        string `json:"sid"`
	Sub        string `json:"sub"`
}

func NewVippsProvider(ext conf.OAuthProviderConfiguration, scopes string) (OAuthProvider, error) {
	if err := ext.ValidateOAuth(); err != nil {
		return nil, err
	}

	oauthScopes := []string{
		"name",
		"email",
	}

	if scopes != "" {
		oauthScopes = append(oauthScopes, strings.Split(scopes, ",")...)
	}

	apiPath := chooseHost(ext.URL, defaultVippsAPIBase)

	return &vippsProvider{
		Config: &oauth2.Config{
			ClientID:     ext.ClientID[0],
			ClientSecret: ext.Secret,
			Endpoint: oauth2.Endpoint{
				AuthURL:  apiPath + "/access-management-1.0/access/oauth2/auth",
				TokenURL: apiPath + "/access-management-1.0/access/oauth2/token",
			},
			RedirectURL: ext.RedirectURI,
			Scopes:      oauthScopes,
		},
		APIPath: apiPath,
	}, nil
}

func (g vippsProvider) GetOAuthToken(code string) (*oauth2.Token, error) {
	return g.Exchange(context.Background(), code)
}

func (g vippsProvider) GetUserData(ctx context.Context, tok *oauth2.Token) (*UserProvidedData, error) {
	var u vippsUser
	if err := makeRequest(ctx, tok, g.Config, g.APIPath+"/vipps-userinfo-api/userinfo", &u); err != nil {
		return nil, err
	}

	data := &UserProvidedData{
		Metadata: &Claims{
			Issuer:            "vipps",
			Subject:           u.Sub,
			Name:              u.Name,
			PreferredUsername: u.Name,
			FullName:          u.Name,
			Email:             strings.ToLower(u.Email),
		},
		Emails: []Email{{
			Email:    u.Email,
			Primary:  true,
			Verified: true,
		}},
	}

	return data, nil
}
