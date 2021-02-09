package provider

import "github.com/allcloud-io/clisso/platform/aws"

const (
	// Okta indicates a Provider of type Okta.
	Okta = "okta"
	// OneLogin indicates a Provider of type OneLogin.
	OneLogin = "onelogin"
)

// Provider represents an identity provider.
type Provider interface {
	GenerateSAMLAssertion() (SAMLAssertion, error)
	// TODO Temporary! Should be broken down into separate functions for generating a SAML
	// assertion, obtaining credentials etc.
	Get(user string, pass string, app App, duration int64) (*aws.Credentials, error)
	Type() string
	Username() string
}

// SAMLAssertion represents a SAML assertion.
type SAMLAssertion struct {
	Raw string
}

func NewSAMLAssertion(s string) SAMLAssertion {
	return SAMLAssertion{Raw: s}
}
