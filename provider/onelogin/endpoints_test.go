package onelogin

import (
	"net/url"
	"testing"
)

func TestEndpoints_SetBase(t *testing.T) {
	for _, test := range []struct {
		name             string
		region           string
		expectVerifyPath string
		expectError      bool
	}{
		{"Region US", "US", "https://api.us.onelogin.com/api/1/saml_assertion/verify_factor", false},
		{"Region EU", "EU", "https://api.eu.onelogin.com/api/1/saml_assertion/verify_factor", false},
		{"No such region", "no such", "", true},
	} {
		t.Run(test.name, func(t *testing.T) {
			e := Endpoints{Region: test.region}

			err := e.setBase()
			if test.expectError && err == nil {
				t.Errorf("expected error")
			}

			if !test.expectError && err != nil {
				t.Errorf("unexpected error %+v", err)
			}

			if test.expectVerifyPath != e.VerifyFactor() {
				t.Errorf("expected %q, received %q", test.expectVerifyPath, e.VerifyFactor())
			}

		})
	}
}

func TestEndpoints_GetUserByEmail(t *testing.T) {
	for _, test := range []struct {
		name    string
		baseURL string
		email   string
		expect  string
	}{
		{"Happy path", "http://example.com", "root@example.com", "http://example.com/api/1/users%3Femail=%25s?email=root%40example.com"},
		{"Empty email", "http://example.com", "", "http://example.com/api/1/users%3Femail=%25s?email="},
	} {
		t.Run(test.name, func(t *testing.T) {
			e := Endpoints{}
			e.base, _ = url.Parse(test.baseURL)

			u := e.GetUserByEmail(test.email)
			if test.expect != u {
				t.Errorf("expected %q, received %q", test.expect, u)
			}
		})
	}
}
