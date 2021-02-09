package keychain

const (
	// KeyChainName is the name of the keychain used to store
	// passwords
	KeyChainName = "clisso"
)

// Keychain provides an interface to allow for the easy testing
// of this package
type Keychain interface {
	Get(string) ([]byte, error)
	Set(string, []byte) error
}

// DefaultKeychain provides a wrapper around github.com/tmc/keyring
// and provides defaults and abstractions for clisso to get passwords
type DefaultKeychain struct{}

// Set takes a provider in an argument, and a password from STDIN, and
// sets it in a keychain, should one exist.
func (DefaultKeychain) Set(provider string, password []byte) (err error) {
	return set(provider, password)
}

// Get will, once given a valid provider, return the password associated
// in order for logins to happen.
// If any error occours while talking to the keychain provider, we silently swallow it
// and just ask the user for the password instead. Error could be anything from access denied to
// password not found.
// TODO: Move password prompt out of this function.
func (DefaultKeychain) Get(provider string) (pw []byte, err error) {
	pass, err := get(provider)
	if err != nil {
		return nil, err
	}
	return pass, nil
}
