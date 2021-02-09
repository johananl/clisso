package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// AppDuration returns the configured duration for the given app.
func AppDuration(app string) int64 {
	return viper.GetInt64(fmt.Sprintf("apps.%s.duration", app))
}

// ProviderDuration returns the configured duration for the given provider.
func ProviderDuration(provider string) int64 {
	return viper.GetInt64(fmt.Sprintf("providers.%s.duration", provider))
}

// ProviderForApp returns the name of the provider for the given app.
func ProviderForApp(app string) string {
	return viper.GetString(fmt.Sprintf("apps.%s.provider", app))
}

// ProviderType returns the type of the given provider.
func ProviderType(provider string) string {
	return viper.GetString(fmt.Sprintf("providers.%s.type", provider))
}

// SelectedApp returns the currently-selected app or an empty string if no app is selected.
func SelectedApp() string {
	return viper.GetString("global.selected-app")
}
