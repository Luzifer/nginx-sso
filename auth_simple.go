package main

import (
	"net/http"

	"github.com/hashicorp/hcl"
)

func init() {
	registerAuthenticator(&authSimple{})
}

type authSimple struct {
	Users  map[string]string   `hcl:"users"`
	Groups map[string][]string `hcl:"groups"`
}

// AuthenticatorID needs to return an unique string to identify
// this special authenticator
func (a authSimple) AuthenticatorID() (id string) { return "simple" }

// Configure loads the configuration for the Authenticator from the
// global config.hcl file which is passed as a byte-slice.
// If no configuration for the Authenticator is supplied the function
// needs to return the AuthenticatorUnconfiguredError
func (a *authSimple) Configure(hclSource []byte) (err error) {
	envelope := struct {
		Providers struct {
			Simple *authSimple `hcl:"simple"`
		} `hcl:"providers"`
	}{}

	if err := hcl.Unmarshal(hclSource, &envelope); err != nil {
		return err
	}

	if envelope.Providers.Simple == nil {
		return authenticatorUnconfiguredError
	}

	a.Users = envelope.Providers.Simple.Users
	a.Groups = envelope.Providers.Simple.Groups

	return nil
}

// DetectUser is used to detect a user without a login form from
// a cookie, header or other methods
// If no user was detected the NoValidUserFoundError needs to be
// returned
func (a authSimple) DetectUser() (user string, groups []string, err error) {
	return "", nil, noValidUserFoundError
}

// Login is called when the user submits the login form and needs
// to authenticate the user or throw an error. If the user has
// successfully logged in the persistent cookie should be written
// in order to use DetectUser for the next login.
// If the user did not login correctly the NoValidUserFoundError
// needs to be returned
func (a authSimple) Login(res http.ResponseWriter, r *http.Request) (err error) { return nil }

// LoginFields needs to return the fields required for this login
// method. If no login using this method is possible the function
// needs to return nil.
func (a authSimple) LoginFields() (fields []loginField) { return nil }

// Logout is called when the user visits the logout endpoint and
// needs to destroy any persistent stored cookies
func (a authSimple) Logout(res http.ResponseWriter) (err error) { return nil }
