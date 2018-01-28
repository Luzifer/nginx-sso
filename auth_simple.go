package main

import (
	"net/http"
	"strings"

	"github.com/Luzifer/go_helpers/str"
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
func (a authSimple) AuthenticatorID() string { return "simple" }

// Configure loads the configuration for the Authenticator from the
// global config.hcl file which is passed as a byte-slice.
// If no configuration for the Authenticator is supplied the function
// needs to return the authenticatorUnconfiguredError
func (a *authSimple) Configure(hclSource []byte) error {
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
// If no user was detected the noValidUserFoundError needs to be
// returned
func (a authSimple) DetectUser(r *http.Request) (string, []string, error) {
	sess, err := cookieStore.Get(r, strings.Join([]string{mainCfg.Cookie.Prefix, a.AuthenticatorID()}, "-"))
	if err != nil {
		return "", nil, noValidUserFoundError
	}

	user, ok := sess.Values["user"].(string)
	if !ok {
		return "", nil, noValidUserFoundError
	}

	groups := []string{}
	for group, users := range a.Groups {
		if str.StringInSlice(user, users) {
			groups = append(groups, group)
		}
	}

	return user, groups, nil
}

// Login is called when the user submits the login form and needs
// to authenticate the user or throw an error. If the user has
// successfully logged in the persistent cookie should be written
// in order to use DetectUser for the next login.
// If the user did not login correctly the noValidUserFoundError
// needs to be returned
func (a authSimple) Login(res http.ResponseWriter, r *http.Request) (err error) { return nil }

// LoginFields needs to return the fields required for this login
// method. If no login using this method is possible the function
// needs to return nil.
func (a authSimple) LoginFields() (fields []loginField) {
	return []loginField{
		{
			Label:       "Username",
			Name:        "username",
			Placeholder: "Username",
			Type:        "text",
		},
		{
			Label:       "Password",
			Name:        "password",
			Placeholder: "****",
			Type:        "password",
		},
	}
}

// Logout is called when the user visits the logout endpoint and
// needs to destroy any persistent stored cookies
func (a authSimple) Logout(res http.ResponseWriter) (err error) { return nil }
