package oidc

import (
	"context"
	"encoding/gob"
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/oauth2"
	yaml "gopkg.in/yaml.v2"

	"github.com/coreos/go-oidc"
	"github.com/gorilla/sessions"
	"github.com/pkg/errors"

	"github.com/Luzifer/nginx-sso/plugins"
)

const (
	userIDMethodFullEmail = "full-email"
	userIDMethodLocalPart = "local-part"
	userIDMethodSubject   = "subject"
)

type AuthOIDC struct {
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
	IssuerName   string `yaml:"issuer_name"`
	IssuerURL    string `yaml:"issuer_url"`
	RedirectURL  string `yaml:"redirect_url"`

	RequireDomain string `yaml:"require_domain"`
	UserIDMethod  string `yaml:"user_id_method"`

	cookie      plugins.CookieConfig
	cookieStore *sessions.CookieStore

	provider *oidc.Provider
}

func init() {
	gob.Register(&oauth2.Token{})
}

func New(cs *sessions.CookieStore) *AuthOIDC {
	return &AuthOIDC{
		IssuerName:   "OpenID Connect",
		UserIDMethod: userIDMethodSubject,
		cookieStore:  cs,
	}
}

// AuthenticatorID needs to return an unique string to identify
// this special authenticator
func (a *AuthOIDC) AuthenticatorID() (id string) { return "oidc" }

// Configure loads the configuration for the Authenticator from the
// global config.yaml file which is passed as a byte-slice.
// If no configuration for the Authenticator is supplied the function
// needs to return the ErrProviderUnconfigured
func (a *AuthOIDC) Configure(yamlSource []byte) (err error) {
	envelope := struct {
		Cookie    plugins.CookieConfig `yaml:"cookie"`
		Providers struct {
			OIDC *AuthOIDC `yaml:"oidc"`
		} `yaml:"providers"`
	}{}

	envelope.Cookie = plugins.DefaultCookieConfig()

	if err := yaml.Unmarshal(yamlSource, &envelope); err != nil {
		return err
	}

	if envelope.Providers.OIDC == nil {
		return plugins.ErrProviderUnconfigured
	}

	a.ClientID = envelope.Providers.OIDC.ClientID
	a.ClientSecret = envelope.Providers.OIDC.ClientSecret
	a.IssuerURL = envelope.Providers.OIDC.IssuerURL
	a.RedirectURL = envelope.Providers.OIDC.RedirectURL
	a.RequireDomain = envelope.Providers.OIDC.RequireDomain

	if envelope.Providers.OIDC.IssuerName != "" {
		a.IssuerName = envelope.Providers.OIDC.IssuerName
	}

	if envelope.Providers.OIDC.UserIDMethod != "" {
		a.UserIDMethod = envelope.Providers.OIDC.UserIDMethod
	}

	a.cookie = envelope.Cookie

	provider, err := oidc.NewProvider(context.Background(), a.IssuerURL)
	if err != nil {
		return errors.Wrap(err, "Unable to fetch provider configuration")
	}
	a.provider = provider

	return nil
}

// DetectUser is used to detect a user without a login form from
// a cookie, header or other methods
// If no user was detected the ErrNoValidUserFound needs to be
// returned
func (a *AuthOIDC) DetectUser(res http.ResponseWriter, r *http.Request) (user string, groups []string, err error) {
	sess, err := a.cookieStore.Get(r, strings.Join([]string{a.cookie.Prefix, a.AuthenticatorID()}, "-"))
	if err != nil {
		return "", nil, plugins.ErrNoValidUserFound
	}

	token, ok := sess.Values["oauth_token"].(*oauth2.Token)
	if !ok {
		return "", nil, plugins.ErrNoValidUserFound
	}

	u, err := a.getUserFromToken(r.Context(), token)
	if err != nil {
		if err == plugins.ErrNoValidUserFound {
			return "", nil, err
		}
		return "", nil, errors.Wrap(err, "Unable to fetch user info")
	}

	// We had a cookie, lets renew it
	sess.Options = a.cookie.GetSessionOpts()
	if err := sess.Save(r, res); err != nil {
		return "", nil, err
	}

	return u, nil, nil // TODO: Maybe get group info?
}

// Login is called when the user submits the login form and needs
// to authenticate the user or throw an error. If the user has
// successfully logged in the persistent cookie should be written
// in order to use DetectUser for the next login.
// With the login result an array of mfaConfig must be returned. In
// case there is no MFA config or the provider does not support MFA
// return nil.
// If the user did not login correctly the ErrNoValidUserFound
// needs to be returned
func (a *AuthOIDC) Login(res http.ResponseWriter, r *http.Request) (user string, mfaConfigs []plugins.MFAConfig, err error) {
	var (
		code  = r.URL.Query().Get("code")
		state = r.URL.Query().Get("state")
		u     string
	)

	if code == "" || state != a.AuthenticatorID() {
		return "", nil, plugins.ErrNoValidUserFound
	}

	token, err := a.getOAuthConfig().Exchange(r.Context(), code)
	if err != nil {
		return "", nil, errors.Wrap(err, "Unable to exchange token")
	}

	u, err = a.getUserFromToken(r.Context(), token)
	if err != nil {
		if err == plugins.ErrNoValidUserFound {
			return "", nil, err
		}
		return "", nil, errors.Wrap(err, "Unable to fetch user info")
	}

	sess, _ := a.cookieStore.Get(r, strings.Join([]string{a.cookie.Prefix, a.AuthenticatorID()}, "-")) // #nosec G104 - On error empty session is returned
	sess.Options = a.cookie.GetSessionOpts()
	sess.Values["oauth_token"] = token

	return u, nil, sess.Save(r, res)
}

// LoginFields needs to return the fields required for this login
// method. If no login using this method is possible the function
// needs to return nil.
func (a *AuthOIDC) LoginFields() (fields []plugins.LoginField) {
	loginURL := a.getOAuthConfig().AuthCodeURL(a.AuthenticatorID())

	return []plugins.LoginField{
		{
			Action:      fmt.Sprintf("window.location.href='%s'", loginURL),
			Label:       "Trigger Login",
			Name:        "button",
			Placeholder: fmt.Sprintf("Sign in with %s", a.IssuerName),
			Type:        "button",
		},
	}
}

// Logout is called when the user visits the logout endpoint and
// needs to destroy any persistent stored cookies
func (a *AuthOIDC) Logout(res http.ResponseWriter, r *http.Request) (err error) {
	sess, _ := a.cookieStore.Get(r, strings.Join([]string{a.cookie.Prefix, a.AuthenticatorID()}, "-")) // #nosec G104 - On error empty session is returned
	sess.Options = a.cookie.GetSessionOpts()
	sess.Options.MaxAge = -1 // Instant delete
	return sess.Save(r, res)
}

// SupportsMFA returns the MFA detection capabilities of the login
// provider. If the provider can provide mfaConfig objects from its
// configuration return true. If this is true the login interface
// will display an additional field for this provider for the user
// to fill in their MFA token.
func (a *AuthOIDC) SupportsMFA() bool { return false }

func (a *AuthOIDC) getOAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     a.ClientID,
		ClientSecret: a.ClientSecret,
		Endpoint:     a.provider.Endpoint(),
		RedirectURL:  a.RedirectURL,
		Scopes: []string{
			oidc.ScopeOpenID,
			"profile",
			"email",
		},
	}
}

func (a *AuthOIDC) getUserFromToken(ctx context.Context, token *oauth2.Token) (string, error) {
	ui, err := a.provider.UserInfo(ctx, oauth2.StaticTokenSource(token))
	if err != nil {
		return "", errors.Wrap(err, "Unable to fetch user info")
	}

	if a.RequireDomain != "" && !strings.HasSuffix(ui.Email, "@"+a.RequireDomain) {
		// E-Mail domain is enforced, ignore all other users
		return "", plugins.ErrNoValidUserFound
	}

	switch a.UserIDMethod {
	case userIDMethodFullEmail:
		return ui.Email, nil

	case userIDMethodLocalPart:
		return strings.Split(ui.Email, "@")[0], nil

	case "":
		fallthrough
	case userIDMethodSubject:
		return ui.Subject, nil

	default:
		return "", errors.Errorf("Invalid user_id_method %q", a.UserIDMethod)
	}
}
