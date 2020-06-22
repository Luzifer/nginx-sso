package google

import (
	"context"
	"encoding/gob"
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	v2 "google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"
	yaml "gopkg.in/yaml.v2"

	"github.com/gorilla/sessions"
	"github.com/pkg/errors"

	"github.com/Luzifer/go_helpers/v2/str"
	"github.com/Luzifer/nginx-sso/plugins"
)

const (
	userIDMethodFullEmail = "full-email"
	userIDMethodLocalPart = "local-part"
	userIDMethodUserID    = "user-id"
)

type AuthGoogleOAuth struct {
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
	RedirectURL  string `yaml:"redirect_url"`

	RequireDomain  string   `yaml:"require_domain"` // Deprecated: Use RequireDomains
	RequireDomains []string `yaml:"require_domains"`
	UserIDMethod   string   `yaml:"user_id_method"`

	cookie      plugins.CookieConfig
	cookieStore *sessions.CookieStore
}

func init() {
	gob.Register(&oauth2.Token{})
}

func New(cs *sessions.CookieStore) *AuthGoogleOAuth {
	return &AuthGoogleOAuth{
		UserIDMethod: userIDMethodUserID,
		cookieStore:  cs,
	}
}

// AuthenticatorID needs to return an unique string to identify
// this special authenticator
func (a *AuthGoogleOAuth) AuthenticatorID() (id string) { return "google_oauth" }

// Configure loads the configuration for the Authenticator from the
// global config.yaml file which is passed as a byte-slice.
// If no configuration for the Authenticator is supplied the function
// needs to return the ErrProviderUnconfigured
func (a *AuthGoogleOAuth) Configure(yamlSource []byte) (err error) {
	envelope := struct {
		Cookie    plugins.CookieConfig `yaml:"cookie"`
		Providers struct {
			GoogleOAuth *AuthGoogleOAuth `yaml:"google_oauth"`
		} `yaml:"providers"`
	}{}

	envelope.Cookie = plugins.DefaultCookieConfig()

	if err := yaml.Unmarshal(yamlSource, &envelope); err != nil {
		return err
	}

	if envelope.Providers.GoogleOAuth == nil {
		return plugins.ErrProviderUnconfigured
	}

	a.ClientID = envelope.Providers.GoogleOAuth.ClientID
	a.ClientSecret = envelope.Providers.GoogleOAuth.ClientSecret
	a.RedirectURL = envelope.Providers.GoogleOAuth.RedirectURL
	a.RequireDomains = envelope.Providers.GoogleOAuth.RequireDomains

	if len(envelope.Providers.GoogleOAuth.RequireDomain) > 0 {
		// Migration for old configuration with only single require_domain
		a.RequireDomains = append(
			a.RequireDomains,
			envelope.Providers.GoogleOAuth.RequireDomain,
		)
	}

	if envelope.Providers.GoogleOAuth.UserIDMethod != "" {
		a.UserIDMethod = envelope.Providers.GoogleOAuth.UserIDMethod
	}

	a.cookie = envelope.Cookie

	return nil
}

// DetectUser is used to detect a user without a login form from
// a cookie, header or other methods
// If no user was detected the ErrNoValidUserFound needs to be
// returned
func (a *AuthGoogleOAuth) DetectUser(res http.ResponseWriter, r *http.Request) (user string, groups []string, err error) {
	sess, err := a.cookieStore.Get(r, strings.Join([]string{a.cookie.Prefix, a.AuthenticatorID()}, "-"))
	if err != nil {
		return "", nil, plugins.ErrNoValidUserFound
	}

	token, ok := sess.Values["google_token"].(*oauth2.Token)
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
func (a *AuthGoogleOAuth) Login(res http.ResponseWriter, r *http.Request) (user string, mfaConfigs []plugins.MFAConfig, err error) {
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
	sess.Values["google_token"] = token

	return u, nil, sess.Save(r, res)
}

// LoginFields needs to return the fields required for this login
// method. If no login using this method is possible the function
// needs to return nil.
func (a *AuthGoogleOAuth) LoginFields() (fields []plugins.LoginField) {
	loginURL := a.getOAuthConfig().AuthCodeURL(
		a.AuthenticatorID(),
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("prompt", "consent"),
	)

	return []plugins.LoginField{
		{
			Action:      fmt.Sprintf("window.location.href='%s'", loginURL),
			Label:       "Trigger Login",
			Name:        "button",
			Placeholder: "Sign in with Google",
			Type:        "button",
		},
	}
}

// Logout is called when the user visits the logout endpoint and
// needs to destroy any persistent stored cookies
func (a *AuthGoogleOAuth) Logout(res http.ResponseWriter, r *http.Request) (err error) {
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
func (a *AuthGoogleOAuth) SupportsMFA() bool { return false }

func (a *AuthGoogleOAuth) getOAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     a.ClientID,
		ClientSecret: a.ClientSecret,
		Endpoint:     google.Endpoint,
		RedirectURL:  a.RedirectURL,
		Scopes: []string{
			v2.UserinfoEmailScope,
			v2.UserinfoProfileScope,
		},
	}
}

func (a *AuthGoogleOAuth) getUserFromToken(ctx context.Context, token *oauth2.Token) (string, error) {
	conf := a.getOAuthConfig()

	httpClient := conf.Client(ctx, token)
	client, err := v2.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return "", errors.Wrap(err, "Unable to instantiate OAuth2 API service")
	}

	tok, err := client.Tokeninfo().Context(ctx).Do()
	if err != nil {
		return "", errors.Wrap(err, "Unable to fetch token-info")
	}

	var mailParts = strings.Split(tok.Email, "@")
	if len(mailParts) != 2 {
		return "", errors.New("Invalid email returned")
	}

	if len(a.RequireDomains) > 0 && !str.StringInSlice(mailParts[1], a.RequireDomains) {
		// E-Mail domain is enforced, ignore all other users
		return "", plugins.ErrNoValidUserFound
	}

	switch a.UserIDMethod {
	case userIDMethodFullEmail:
		return tok.Email, nil

	case userIDMethodLocalPart:
		return strings.Split(tok.Email, "@")[0], nil

	case "":
		fallthrough
	case userIDMethodUserID:
		return tok.UserId, nil

	default:
		return "", errors.Errorf("Invalid user_id_method %q", a.UserIDMethod)
	}
}
