package main

import (
	"net/http"
	"strings"

	crowd "github.com/jda/go-crowd"
	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
)

func init() {
	registerAuthenticator(&authCrowd{})
}

type authCrowd struct {
	URL         string `yaml:"url"`
	AppName     string `yaml:"app_name"`
	AppPassword string `yaml:"app_pass"`

	crowd crowd.Crowd
}

// AuthenticatorID needs to return an unique string to identify
// this special authenticator
func (a authCrowd) AuthenticatorID() string { return "crowd" }

// Configure loads the configuration for the Authenticator from the
// global config.yaml file which is passed as a byte-slice.
// If no configuration for the Authenticator is supplied the function
// needs to return the errProviderUnconfigured
func (a *authCrowd) Configure(yamlSource []byte) error {
	envelope := struct {
		Providers struct {
			Crowd *authCrowd `yaml:"crowd"`
		} `yaml:"providers"`
	}{}

	if err := yaml.Unmarshal(yamlSource, &envelope); err != nil {
		return err
	}

	if envelope.Providers.Crowd == nil {
		return errProviderUnconfigured
	}

	a.URL = envelope.Providers.Crowd.URL
	a.AppName = envelope.Providers.Crowd.AppName
	a.AppPassword = envelope.Providers.Crowd.AppPassword

	if a.AppName == "" || a.AppPassword == "" {
		return errProviderUnconfigured
	}

	var err error
	a.crowd, err = crowd.New(a.AppName, a.AppPassword, a.URL)

	return err
}

// DetectUser is used to detect a user without a login form from
// a cookie, header or other methods
// If no user was detected the errNoValidUserFound needs to be
// returned
func (a authCrowd) DetectUser(res http.ResponseWriter, r *http.Request) (string, []string, error) {
	cc, err := a.crowd.GetCookieConfig()
	if err != nil {
		return "", nil, err
	}

	cookie, err := r.Cookie(cc.Name)
	switch err {
	case nil:
		// Fine, we do have a cookie
	case http.ErrNoCookie:
		// Also fine, there is no cookie
		return "", nil, errNoValidUserFound
	default:
		return "", nil, err
	}

	ssoToken := cookie.Value
	sess, err := a.crowd.GetSession(ssoToken)
	if err != nil {
		log.WithError(err).Debug("Getting crowd session failed")
		return "", nil, errNoValidUserFound
	}

	user := sess.User.UserName
	cGroups, err := a.crowd.GetDirectGroups(user)
	if err != nil {
		return "", nil, err
	}

	groups := []string{}
	for _, g := range cGroups {
		groups = append(groups, g.Name)
	}

	return user, groups, nil
}

// Login is called when the user submits the login form and needs
// to authenticate the user or throw an error. If the user has
// successfully logged in the persistent cookie should be written
// in order to use DetectUser for the next login.
// If the user did not login correctly the errNoValidUserFound
// needs to be returned
func (a authCrowd) Login(res http.ResponseWriter, r *http.Request) error {
	username := r.FormValue(strings.Join([]string{a.AuthenticatorID(), "username"}, "-"))
	password := r.FormValue(strings.Join([]string{a.AuthenticatorID(), "password"}, "-"))

	cc, err := a.crowd.GetCookieConfig()
	if err != nil {
		return err
	}

	sess, err := a.crowd.NewSession(username, password, r.RemoteAddr)
	if err != nil {
		log.WithFields(log.Fields{
			"username": username,
		}).WithError(err).Debug("Crowd authentication failed")
		return errNoValidUserFound
	}

	http.SetCookie(res, &http.Cookie{
		Name:     cc.Name,
		Value:    sess.Token,
		Path:     "/",
		Domain:   cc.Domain,
		Expires:  sess.Expires,
		Secure:   cc.Secure,
		HttpOnly: true,
	})

	return nil
}

// LoginFields needs to return the fields required for this login
// method. If no login using this method is possible the function
// needs to return nil.
func (a authCrowd) LoginFields() (fields []loginField) {
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
func (a authCrowd) Logout(res http.ResponseWriter, r *http.Request) (err error) {
	cc, err := a.crowd.GetCookieConfig()
	if err != nil {
		return err
	}

	http.SetCookie(res, &http.Cookie{
		Name:     cc.Name,
		Value:    "",
		Path:     "/",
		Domain:   cc.Domain,
		MaxAge:   -1,
		Secure:   cc.Secure,
		HttpOnly: true,
	})

	return nil
}
