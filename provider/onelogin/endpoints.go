package onelogin

import (
	"fmt"
	"net/url"
)

const (
	usBase = "https://api.us.onelogin.com"
	euBase = "https://api.eu.onelogin.com"

	GenerateSamlAssertionPath string = "/api/1/saml_assertion"
	GenerateTokensPath        string = "/auth/oauth2/v2/token"
	GetUserByEmailPath        string = "/api/1/users?email=%s"
	VerifyFactorPath          string = "/api/1/saml_assertion/verify_factor"
)

// Endpoints represent the OneLogin API HTTP endpoints.
type Endpoints struct {
	Region string

	base *url.URL
}

func (e *Endpoints) setBase() (err error) {
	var base string
	switch e.Region {
	case "US":
		base = usBase

	case "EU":
		base = euBase

	default:
		return fmt.Errorf("Region %q is an invalid OneLogin region. Valid values are EU or US.", e.Region)
	}

	e.base, err = url.Parse(base)

	return
}

// GenerateSamlAssertion will return a the relevant Generate SAML Assertion
// endpoint for a given base URL
func (e Endpoints) GenerateSamlAssertion() string {
	return e.doURL(GenerateSamlAssertionPath, make(url.Values))
}

// GenerateTokens will return the relevant Generate Tokens endpoint
// for a base URL
func (e Endpoints) GenerateTokens() string {
	return e.doURL(GenerateTokensPath, make(url.Values))
}

// GetUserByEmail will, given an email address, return a valid url
// to search the Users endpoint by email address
func (e Endpoints) GetUserByEmail(email string) string {
	return e.doURL(GetUserByEmailPath, url.Values{"email": []string{email}})
}

// VerifyFactor will return a valid URL for requests to check MFA tokens
func (e Endpoints) VerifyFactor() string {
	return e.doURL(VerifyFactorPath, make(url.Values))
}

func (e Endpoints) doURL(endpoint string, params url.Values) string {
	if e.base == nil {
		return ""
	}

	e.base.Path = endpoint
	e.base.RawQuery = params.Encode()

	return e.base.String()
}
