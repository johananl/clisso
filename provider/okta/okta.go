package okta

import (
	"fmt"

	"github.com/allcloud-io/clisso/provider"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

// ProviderConfig represents an Okta provider configuration.
type ProviderConfig struct {
	BaseURL  string
	Username string
}

// NewProviderConfig reads the configuration for provider p and returns a pointer to a
// ProviderConfig struct.
func NewProviderConfig(p string) (*ProviderConfig, error) {
	baseURL := viper.GetString(fmt.Sprintf("providers.%s.base-url", p))
	username := viper.GetString(fmt.Sprintf("providers.%s.username", p))

	if baseURL == "" {
		return nil, errors.New("base-url config value must bet set")
	}

	return &ProviderConfig{BaseURL: baseURL, Username: username}, nil
}

// Provider is a Provider implementation for Okta.
type Provider struct {
	Client *Client
	Config *ProviderConfig
	Name   string
}

func (p *Provider) GenerateSAMLAssertion() (provider.SAMLAssertion, error) {
	// TODO
	return provider.NewSAMLAssertion("fake"), nil
}

func (p *Provider) Type() string {
	return "Okta"
}

func (p *Provider) Username() string {
	return p.Config.Username
}

// New constructs a new OktaProvider and returns a pointer to it.
func New(name string, pc *ProviderConfig) (*Provider, error) {
	c, err := NewClient(pc.BaseURL)
	if err != nil {
		return nil, errors.Wrap(err, "creating Okta client")
	}

	return &Provider{
		Client: c,
		Config: pc,
		Name:   name,
	}, nil
}

// App represents an app configured on Okta.
type App struct {
	// Okta app URL.
	url string
}

// ID returns the Okta app's URL.
func (a *App) ID() string {
	return a.url
}

// NewApp constructs a new App and returns a pointer to it.
func NewApp(name string) (*App, error) {
	url := viper.GetString(fmt.Sprintf("apps.%s.url", name))
	if url == "" {
		return nil, fmt.Errorf("'url' config parameter not set for app '%s'", name)
	}

	return &App{url: url}, nil
}
