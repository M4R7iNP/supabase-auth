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

type vippsAddress struct {
	AddressType   string `json:"address_type"`
	Country       string `json:"country"`
	Formatted     string `json:"formatted"`
	PostalCode    string `json:"postal_code"`
	Region        string `json:"region"`
	StreetAddress string `json:"street_address"`
}

type vippsDelegatedConsent struct {
	Language             string `json:"language"`
	Heading              string `json:"heading"`
	TermsDescription     string `json:"termsDescription"`
	ConfirmConsentButton string `json:"confirmConsentButtonText"`
	Links                struct {
		TermsLinkText            string `json:"termsLinkText"`
		TermsLinkURL             string `json:"termsLinkUrl"`
		PrivacyStatementLinkText string `json:"privacyStatementLinkText"`
		PrivacyStatementLinkURL  string `json:"privacyStatementLinkUrl"`
	} `json:"links"`
	TimeOfConsent string `json:"timeOfConsent"`
	Consents      []struct {
		ID                  string `json:"id"`
		Accepted            bool   `json:"accepted"`
		Required            bool   `json:"required"`
		TextDisplayedToUser string `json:"textDisplayedToUser"`
	} `json:"consents"`
}

type vippsAgreement struct {
	Language    string `json:"language"`
	Description string `json:"description"`
	ButtonText  string `json:"buttonText"`
	Links       struct {
		TermsLinkText            string `json:"termsLinkText"`
		TermsLinkURL             string `json:"termsLinkUrl"`
		PrivacyStatementLinkText string `json:"privacyStatementLinkText"`
		PrivacyStatementLinkURL  string `json:"privacyStatementLinkUrl"`
	} `json:"links"`
	Timestamp string `json:"timestamp"`
	Accepted  bool   `json:"accepted"`
}

// https://developer.vippsmobilepay.com/api/login/#operation/userinfoAuthorizationCode
type vippsUser struct {
	FamilyName          string                  `json:"family_name"`
	GivenName           string                  `json:"given_name"`
	Name                string                  `json:"name"`
	Email               string                  `json:"email"`
	EmailVerified       bool                    `json:"email_verified"`
	Sid                 string                  `json:"sid"`
	Sub                 string                  `json:"sub"`
	Birthdate           string                  `json:"birthdate"`
	PhoneNumber         string                  `json:"phone_number"`
	PhoneNumberVerified bool                    `json:"phone_number_verified"`
	Gender              string                  `json:"gender"`
	Address             []vippsAddress          `json:"address"`
	OtherAddresses      []vippsAddress          `json:"other_addresses"`
	DelegatedConsents   []vippsDelegatedConsent `json:"delegated_consents"`
	Agreement           []vippsAgreement        `json:"agreement"`
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
			EmailVerified:     u.EmailVerified,
			Birthdate:         u.Birthdate,
			Phone:             u.PhoneNumber,
			PhoneVerified:     u.PhoneNumberVerified,
			Gender:            u.Gender,
			CustomClaims: map[string]any{
				"address":            u.Address,
				"other_addresses":    u.OtherAddresses,
				"delegated_consents": u.DelegatedConsents,
				"agreement":          u.Agreement,
			},
		},
		Emails: []Email{{
			Email:    u.Email,
			Primary:  true,
			Verified: true,
		}},
	}

	return data, nil
}
