package main

import (
	"net/http"
	"strings"

	"github.com/GeertJohan/yubigo"
	"github.com/Luzifer/go_helpers/str"
	yaml "gopkg.in/yaml.v2"
)

func init() {
	registerAuthenticator(&authYubikey{})
}

type authYubikey struct {
	ClientID  string              `yaml:"client_id"`
	SecretKey string              `yaml:"secret_key"`
	Devices   map[string]string   `yaml:"devices"`
	Groups    map[string][]string `yaml:"groups"`
}

// AuthenticatorID needs to return an unique string to identify
// this special authenticator
func (a authYubikey) AuthenticatorID() string { return "yubikey" }

// Configure loads the configuration for the Authenticator from the
// global config.yaml file which is passed as a byte-slice.
// If no configuration for the Authenticator is supplied the function
// needs to return the errAuthenticatorUnconfigured
func (a *authYubikey) Configure(yamlSource []byte) error {
	envelope := struct {
		Providers struct {
			Yubikey *authYubikey `yaml:"yubikey"`
		} `yaml:"providers"`
	}{}

	if err := yaml.Unmarshal(yamlSource, &envelope); err != nil {
		return err
	}

	if envelope.Providers.Yubikey == nil {
		return errAuthenticatorUnconfigured
	}

	a.ClientID = envelope.Providers.Yubikey.ClientID
	a.SecretKey = envelope.Providers.Yubikey.SecretKey
	a.Devices = envelope.Providers.Yubikey.Devices
	a.Groups = envelope.Providers.Yubikey.Groups

	return nil
}

// DetectUser is used to detect a user without a login form from
// a cookie, header or other methods
// If no user was detected the errNoValidUserFound needs to be
// returned
func (a authYubikey) DetectUser(res http.ResponseWriter, r *http.Request) (string, []string, error) {
	sess, err := cookieStore.Get(r, strings.Join([]string{mainCfg.Cookie.Prefix, a.AuthenticatorID()}, "-"))
	if err != nil {
		return "", nil, errNoValidUserFound
	}

	user, ok := sess.Values["user"].(string)
	if !ok {
		return "", nil, errNoValidUserFound
	}

	// We had a cookie, lets renew it
	sess.Options = mainCfg.GetSessionOpts()
	if err := sess.Save(r, res); err != nil {
		return "", nil, err
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
// If the user did not login correctly the errNoValidUserFound
// needs to be returned
func (a authYubikey) Login(res http.ResponseWriter, r *http.Request) error {
	keyInput := r.FormValue(strings.Join([]string{a.AuthenticatorID(), "key-input"}, "-"))

	yubiAuth, err := yubigo.NewYubiAuth(a.ClientID, a.SecretKey)
	if err != nil {
		return err
	}

	_, ok, err := yubiAuth.Verify(keyInput)
	if err != nil {
		return err
	}

	if !ok {
		// Not a valid authentication
		return errNoValidUserFound
	}

	user, ok := a.Devices[keyInput[:12]]
	if !ok {
		// We do not have a definition for that key
		return errNoValidUserFound
	}

	sess, _ := cookieStore.Get(r, strings.Join([]string{mainCfg.Cookie.Prefix, a.AuthenticatorID()}, "-"))
	sess.Options = mainCfg.GetSessionOpts()
	sess.Values["user"] = user
	return sess.Save(r, res)
}

// LoginFields needs to return the fields required for this login
// method. If no login using this method is possible the function
// needs to return nil.
func (a authYubikey) LoginFields() (fields []loginField) {
	return []loginField{
		{
			Label:       "Yubikey One-Time-Password",
			Name:        "key-input",
			Placeholder: "Press the button of your Yubikey...",
			Type:        "text",
		},
	}
}

// Logout is called when the user visits the logout endpoint and
// needs to destroy any persistent stored cookies
func (a authYubikey) Logout(res http.ResponseWriter, r *http.Request) (err error) {
	sess, _ := cookieStore.Get(r, strings.Join([]string{mainCfg.Cookie.Prefix, a.AuthenticatorID()}, "-"))
	sess.Options = mainCfg.GetSessionOpts()
	sess.Options.MaxAge = -1 // Instant delete
	return sess.Save(r, res)
}
