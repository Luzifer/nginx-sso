package main

import (
	"net/http"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

func init() {
	registerAuthenticator(&authToken{})
}

type authToken struct {
	Tokens map[string]string `yaml:"tokens"`
}

// AuthenticatorID needs to return an unique string to identify
// this special authenticator
func (a authToken) AuthenticatorID() string { return "token" }

// Configure loads the configuration for the Authenticator from the
// global config.yaml file which is passed as a byte-slice.
// If no configuration for the Authenticator is supplied the function
// needs to return the errAuthenticatorUnconfigured
func (a *authToken) Configure(yamlSource []byte) error {
	envelope := struct {
		Providers struct {
			Token *authToken `yaml:"token"`
		} `yaml:"providers"`
	}{}

	if err := yaml.Unmarshal(yamlSource, &envelope); err != nil {
		return err
	}

	if envelope.Providers.Token == nil {
		return errAuthenticatorUnconfigured
	}

	a.Tokens = envelope.Providers.Token.Tokens

	return nil
}

// DetectUser is used to detect a user without a login form from
// a cookie, header or other methods
// If no user was detected the errNoValidUserFound needs to be
// returned
func (a authToken) DetectUser(r *http.Request) (string, []string, error) {
	authHeader := r.Header.Get("Authorization")

	if !strings.HasPrefix(authHeader, "Token ") {
		return "", nil, errNoValidUserFound
	}

	tmp := strings.SplitN(authHeader, " ", 2)
	suppliedToken := tmp[1]

	for user, token := range a.Tokens {
		if token == suppliedToken {
			return user, nil, nil
		}
	}

	return "", nil, errNoValidUserFound
}

// Login is called when the user submits the login form and needs
// to authenticate the user or throw an error. If the user has
// successfully logged in the persistent cookie should be written
// in order to use DetectUser for the next login.
// If the user did not login correctly the errNoValidUserFound
// needs to be returned
func (a authToken) Login(res http.ResponseWriter, r *http.Request) error { return errNoValidUserFound }

// LoginFields needs to return the fields required for this login
// method. If no login using this method is possible the function
// needs to return nil.
func (a authToken) LoginFields() []loginField { return nil }

// Logout is called when the user visits the logout endpoint and
// needs to destroy any persistent stored cookies
func (a authToken) Logout(res http.ResponseWriter, r *http.Request) error { return nil }
