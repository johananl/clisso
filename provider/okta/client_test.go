package okta

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func getTestServer(data string) *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(data))
	}))

	return ts
}

var c = Client{}

func TestGetSessionToken(t *testing.T) {
	data := `{
		"expiresAt": "2018-11-03T10:15:57.000Z",
		"status": "SUCCESS",
		"sessionToken": "fake_token"
	}`

	ts := getTestServer(data)
	defer ts.Close()

	c.BaseURL = ts.URL

	resp, err := c.GetSessionToken(&GetSessionTokenParams{Username: "test", Password: "test"})
	if err != nil {
		t.Errorf("getting session token: %s", err)
	}
	if resp.SessionToken != "fake_token" {
		t.Errorf("Wrong response, got: %v, want: %v", resp.SessionToken, "fake_token")
	}
}

func TestGetSessionTokenMFARequired(t *testing.T) {
	data := `{
		"stateToken": "fake_token",
		"expiresAt": "2018-07-08T13:47:53.000Z",
		"status": "MFA_REQUIRED",
		"_embedded": {
			"factors": [
				{
					"id": "fake_id",
					"factorType": "token:software:totp",
					"provider": "GOOGLE",
					"vendorName": "GOOGLE",
					"profile": {
						"credentialId": "test@test.local"
					},
					"_links": {
						"verify": {
							"href": "https://test.test.local/api/v1/authn/factors/fake/verify",
							"hints": {
								"allow": [
									"POST"
								]
							}
						}
					}
				}
			]
		}
	}`

	ts := getTestServer(data)
	defer ts.Close()

	c.BaseURL = ts.URL

	resp, err := c.GetSessionToken(&GetSessionTokenParams{Username: "test", Password: "test"})
	if err != nil {
		t.Errorf("getting session token: %v", err)
	}
	if resp.StateToken != "fake_token" {
		t.Errorf("Wrong response, got: %v, want: %v", resp.StateToken, "fake_token")
	}
}

func TestVerifyFactor(t *testing.T) {
	data := `{
		"expiresAt": "2015-11-03T10:15:57.000Z",
		"status": "SUCCESS",
		"sessionToken": "fake_token"
	}`

	ts := getTestServer(data)
	defer ts.Close()

	c.BaseURL = ts.URL

	resp, err := c.VerifyFactor(&VerifyFactorParams{
		FactorID:   "fake_id",
		StateToken: "fake_state_token",
		PassCode:   "123456",
	})
	if err != nil {
		t.Errorf("verifying factor: %v", err)
	}

	if resp.Status != "SUCCESS" {
		t.Errorf("Wrong response, got: %v, want: %v", resp.Status, "SUCCESS")
	}
	if resp.SessionToken != "fake_token" {
		t.Errorf("Wrong response, got: %v, want: %v", resp.SessionToken, "fake_token")
	}
	exp, _ := time.Parse("2006-01-02T15:04:05.000Z", "2015-11-03T10:15:57.000Z")
	if resp.ExpiresAt != exp {
		t.Errorf("Wrong response, got: %v, want: %v", resp.ExpiresAt, exp)
	}
}
