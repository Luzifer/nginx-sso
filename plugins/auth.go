package plugins

import "net/http"

type Authenticator interface {
	// AuthenticatorID needs to return an unique string to identify
	// this special authenticator
	AuthenticatorID() (id string)

	// Configure loads the configuration for the Authenticator from the
	// global config.yaml file which is passed as a byte-slice.
	// If no configuration for the Authenticator is supplied the function
	// needs to return the ErrProviderUnconfigured
	Configure(yamlSource []byte) (err error)

	// DetectUser is used to detect a user without a login form from
	// a cookie, header or other methods
	// If no user was detected the ErrNoValidUserFound needs to be
	// returned
	DetectUser(res http.ResponseWriter, r *http.Request) (user string, groups []string, err error)

	// Login is called when the user submits the login form and needs
	// to authenticate the user or throw an error. If the user has
	// successfully logged in the persistent cookie should be written
	// in order to use DetectUser for the next login.
	// With the login result an array of mfaConfig must be returned. In
	// case there is no MFA config or the provider does not support MFA
	// return nil.
	// If the user did not login correctly the ErrNoValidUserFound
	// needs to be returned
	Login(res http.ResponseWriter, r *http.Request) (user string, mfaConfigs []MFAConfig, err error)

	// LoginFields needs to return the fields required for this login
	// method. If no login using this method is possible the function
	// needs to return nil.
	LoginFields() (fields []LoginField)

	// Logout is called when the user visits the logout endpoint and
	// needs to destroy any persistent stored cookies
	Logout(res http.ResponseWriter, r *http.Request) (err error)

	// SupportsMFA returns the MFA detection capabilities of the login
	// provider. If the provider can provide mfaConfig objects from its
	// configuration return true. If this is true the login interface
	// will display an additional field for this provider for the user
	// to fill in their MFA token.
	SupportsMFA() bool
}

type LoginField struct {
	Action      string
	Label       string
	Name        string
	Placeholder string
	Type        string
}
