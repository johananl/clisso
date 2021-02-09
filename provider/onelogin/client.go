package onelogin

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// Client represents a OneLogin API client.
type Client struct {
	http.Client
	Endpoints Endpoints
}

type GenerateTokensParams struct {
	GrantType string `json:"grant_type"`
}

type GenerateTokensResponse struct {
	AccessToken  string    `json:"access_token"`
	CreatedAt    time.Time `json:"created_at"`
	ExpiresIn    int       `json:"expires_in"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	AccountID    int       `json:"account_id"`
}

type GenerateSamlAssertionParams struct {
	UsernameOrEmail string `json:"username_or_email"`
	Password        string `json:"password"`
	AppId           string `json:"app_id"`
	Subdomain       string `json:"subdomain"`
}

// TODO This one assumes MFA is enabled. Need to handle all cases.
type GenerateSamlAssertionResponse struct {
	Status struct {
		Error   bool   `json:"error"`
		Code    int    `json:"code"`
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"status"`
	Data []struct {
		StateToken  string `json:"state_token"`
		Devices     []Device
		CallbackUrl string `json:"callback_url"`
		User        struct {
			Lastname  string `json:"lastname"`
			Username  string `json:"username"`
			Email     string `json:"email"`
			Firstname string `json:"firstname"`
			Id        int    `json:"id"`
		}
	}
}

type VerifyFactorParams struct {
	AppId       string `json:"app_id"`
	DeviceId    string `json:"device_id"`
	StateToken  string `json:"state_token"`
	OtpToken    string `json:"otp_token"`
	DoNotNotify bool   `json:"do_not_notify"`
}

type VerifyFactorResponse struct {
	Status struct {
		Error   bool   `json:"error"`
		Code    int    `json:"code"`
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"status"`
	Data string `json:"data"`
}

type GetUserByEmailResponse struct {
	Status struct {
		Error   bool   `json:"error"`
		Code    int    `json:"code"`
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"status"`
	Data []struct {
		ID int `json:"id"`
	}
}

type Device struct {
	DeviceID   int    `json:"device_id"`
	DeviceType string `json:"device_type"`
}

// makeRequest constructs an HTTP request and returns a pointer to it.
// TODO Wrap arguments in a type
func makeRequest(method string, url string, headers map[string]string, body interface{}) (*http.Request, error) {
	json, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("parsing body: %v", err)
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(json))
	if err != nil {
		return nil, fmt.Errorf("making HTTP request: %v", err)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	return req, nil
}

// doRequest gets a pointer to an HTTP request and an HTTP client, executes the request
// using the client, handles any HTTP-related errors and returns any data as a string.
func (c *Client) doRequest(r *http.Request) (string, error) {
	resp, err := c.Do(r)
	if err != nil {
		return "", fmt.Errorf("sending HTTP request: %v", err)
	}

	if resp.StatusCode != 200 {
		return "", errors.New(resp.Status)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	b := []byte(body)

	return string(b), nil
}

// GenerateTokens generates the tokens required for interacting with the OneLogin
// API.
func (c *Client) GenerateTokens(clientID, clientSecret string) (string, error) {
	headers := map[string]string{
		"Authorization": fmt.Sprintf("client_id:%v, client_secret:%v", clientID, clientSecret),
		"Content-Type":  "application/json",
	}
	body := GenerateTokensParams{GrantType: "client_credentials"}

	req, err := makeRequest(http.MethodPost, c.Endpoints.GenerateTokens(), headers, &body)
	if err != nil {
		return "", fmt.Errorf("creating request: %v", err)
	}

	data, err := c.doRequest(req)
	if err != nil {
		return "", fmt.Errorf("doing HTTP request: %v", err)
	}

	var resp GenerateTokensResponse
	if err := json.Unmarshal([]byte(data), &resp); err != nil {
		return "", fmt.Errorf("parsing HTTP response: %v", err)
	}

	// TODO add handling for valid JSON with wrong response

	return resp.AccessToken, nil
}

// GenerateSamlAssertion gets a OneLogin access token and a GenerateSamlAssertionParams struct
// and returns a GenerateSamlAssertionResponse.
// TODO improve doc
func (c *Client) GenerateSamlAssertion(token string, p *GenerateSamlAssertionParams) (*GenerateSamlAssertionResponse, error) {
	headers := map[string]string{
		"Authorization": fmt.Sprintf("bearer:%v", token),
		"Content-Type":  "application/json",
	}
	body := p

	req, err := makeRequest(http.MethodPost, c.Endpoints.GenerateSamlAssertion(), headers, &body)
	if err != nil {
		return nil, fmt.Errorf("creating request: %v", err)
	}

	data, err := c.doRequest(req)
	// TODO An invalid Onelogin app ID gives HTTP 404 here. Need to show a nice
	// error in this case.
	if err != nil {
		//if oneLoginError, ok := err.(*OneLoginError); ok {
		//	fmt.Println(oneLoginError.StatusCode)
		//}
		return nil, fmt.Errorf("doing HTTP request: %v", err)
	}

	var resp GenerateSamlAssertionResponse
	if err := json.Unmarshal([]byte(data), &resp); err != nil {
		return nil, fmt.Errorf("parsing HTTP response: %v", err)
	}

	return &resp, nil
}

// VerifyFactor gets a OneLogin access token and a VerifyFactorParams struct and returns a
// VerifyFactorResponse.
func (c *Client) VerifyFactor(token string, p *VerifyFactorParams) (*VerifyFactorResponse, error) {
	headers := map[string]string{
		"Authorization": fmt.Sprintf("bearer:%v", token),
		"Content-Type":  "application/json",
	}
	body := p

	req, err := makeRequest(http.MethodPost, c.Endpoints.VerifyFactor(), headers, &body)
	if err != nil {
		// TODO Let the user know which method generated the error
		return nil, fmt.Errorf("creating request: %v", err)
	}

	data, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("doing HTTP request: %v", err)
	}

	var resp VerifyFactorResponse
	if err := json.Unmarshal([]byte(data), &resp); err != nil {
		return nil, fmt.Errorf("parsing HTTP response: %v", err)
	}

	return &resp, nil
}

// NewClient creates a new Client and returns a pointer to it.
func NewClient(region string) (c *Client, err error) {
	c = new(Client)

	c.Endpoints = Endpoints{Region: region}
	err = c.Endpoints.setBase()

	return
}
