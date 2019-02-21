package main

import (
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"
	yaml "gopkg.in/yaml.v2"

	"github.com/Luzifer/go_helpers/str"
	"github.com/Luzifer/nginx-sso/plugins"
)

func init() {
	registerAuthenticator(&authSimple{})
}

type authSimple struct {
	EnableBasicAuth bool                           `yaml:"enable_basic_auth"`
	Users           map[string]string              `yaml:"users"`
	Groups          map[string][]string            `yaml:"groups"`
	MFA             map[string][]plugins.MFAConfig `yaml:"mfa"`
}

// AuthenticatorID needs to return an unique string to identify
// this special authenticator
func (a authSimple) AuthenticatorID() string { return "simple" }

// Configure loads the configuration for the Authenticator from the
// global config.yaml file which is passed as a byte-slice.
// If no configuration for the Authenticator is supplied the function
// needs to return the plugins.ErrProviderUnconfigured
func (a *authSimple) Configure(yamlSource []byte) error {
	envelope := struct {
		Providers struct {
			Simple *authSimple `yaml:"simple"`
		} `yaml:"providers"`
	}{}

	if err := yaml.Unmarshal(yamlSource, &envelope); err != nil {
		return err
	}

	if envelope.Providers.Simple == nil {
		return plugins.ErrProviderUnconfigured
	}

	a.EnableBasicAuth = envelope.Providers.Simple.EnableBasicAuth
	a.Users = envelope.Providers.Simple.Users
	a.Groups = envelope.Providers.Simple.Groups
	a.MFA = envelope.Providers.Simple.MFA

	return nil
}

// DetectUser is used to detect a user without a login form from
// a cookie, header or other methods
// If no user was detected the plugins.ErrNoValidUserFound needs to be
// returned
func (a authSimple) DetectUser(res http.ResponseWriter, r *http.Request) (string, []string, error) {
	var user string

	if a.EnableBasicAuth {
		if basicUser, basicPass, ok := r.BasicAuth(); ok {
			for u, p := range a.Users {
				if u != basicUser {
					continue
				}
				if bcrypt.CompareHashAndPassword([]byte(p), []byte(basicPass)) != nil {
					continue
				}

				user = basicUser
			}
		}
	}

	if user == "" {
		sess, err := cookieStore.Get(r, strings.Join([]string{mainCfg.Cookie.Prefix, a.AuthenticatorID()}, "-"))
		if err != nil {
			return "", nil, plugins.ErrNoValidUserFound
		}

		var ok bool
		user, ok = sess.Values["user"].(string)
		if !ok {
			return "", nil, plugins.ErrNoValidUserFound
		}

		// We had a cookie, lets renew it
		sess.Options = mainCfg.GetSessionOpts()
		if err := sess.Save(r, res); err != nil {
			return "", nil, err
		}
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
// If the user did not login correctly the plugins.ErrNoValidUserFound
// needs to be returned
func (a authSimple) Login(res http.ResponseWriter, r *http.Request) (string, []plugins.MFAConfig, error) {
	username := r.FormValue(strings.Join([]string{a.AuthenticatorID(), "username"}, "-"))
	password := r.FormValue(strings.Join([]string{a.AuthenticatorID(), "password"}, "-"))

	for u, p := range a.Users {
		if u != username {
			continue
		}
		if bcrypt.CompareHashAndPassword([]byte(p), []byte(password)) != nil {
			continue
		}

		sess, _ := cookieStore.Get(r, strings.Join([]string{mainCfg.Cookie.Prefix, a.AuthenticatorID()}, "-")) // #nosec G104 - On error empty session is returned
		sess.Options = mainCfg.GetSessionOpts()
		sess.Values["user"] = u
		return u, a.MFA[u], sess.Save(r, res)
	}

	return "", nil, plugins.ErrNoValidUserFound
}

// LoginFields needs to return the fields required for this login
// method. If no login using this method is possible the function
// needs to return nil.
func (a authSimple) LoginFields() (fields []plugins.LoginField) {
	return []plugins.LoginField{
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
func (a authSimple) Logout(res http.ResponseWriter, r *http.Request) (err error) {
	sess, _ := cookieStore.Get(r, strings.Join([]string{mainCfg.Cookie.Prefix, a.AuthenticatorID()}, "-")) // #nosec G104 - On error empty session is returned
	sess.Options = mainCfg.GetSessionOpts()
	sess.Options.MaxAge = -1 // Instant delete
	return sess.Save(r, res)
}

// SupportsMFA returns the MFA detection capabilities of the login
// provider. If the provider can provide mfaConfig objects from its
// configuration return true. If this is true the login interface
// will display an additional field for this provider for the user
// to fill in their MFA token.
func (a authSimple) SupportsMFA() bool { return true }
