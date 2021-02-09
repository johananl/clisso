package provider

// App represents a configured app on an identity provider.
type App interface {
	// Returns a string which uniquely identifies the app on the identity provider.
	ID() string
}
