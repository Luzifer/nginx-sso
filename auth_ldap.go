package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	ldap "gopkg.in/ldap.v2"
	yaml "gopkg.in/yaml.v2"
)

func init() {
	registerAuthenticator(&authLDAP{})
}

type authLDAP struct {
	EnableBasicAuth       bool   `yaml:"enable_basic_auth"`
	GroupMembershipFilter string `yaml:"group_membership_filter"`
	GroupSearchBase       string `yaml:"group_search_base"`
	ManagerDN             string `yaml:"manager_dn"`
	ManagerPassword       string `yaml:"manager_password"`
	RootDN                string `yaml:"root_dn"`
	Server                string `yaml:"server"`
	UserSearchBase        string `yaml:"user_search_base"`
	UserSearchFilter      string `yaml:"user_search_filter"`
	UsernameAttribute     string `yaml:"username_attribute"`
	TLSConfig             *struct {
		ValidateHostname string `yaml:"validate_hostname"`
		AllowInsecure    bool   `yaml:"allow_insecure"`
	} `yaml:"tls_config"`
}

// AuthenticatorID needs to return an unique string to identify
// this special authenticator
func (a authLDAP) AuthenticatorID() string { return "ldap" }

// Configure loads the configuration for the Authenticator from the
// global config.yaml file which is passed as a byte-slice.
// If no configuration for the Authenticator is supplied the function
// needs to return the errProviderUnconfigured
func (a *authLDAP) Configure(yamlSource []byte) error {
	envelope := struct {
		Providers struct {
			LDAP *authLDAP `yaml:"ldap"`
		} `yaml:"providers"`
	}{}

	if err := yaml.Unmarshal(yamlSource, &envelope); err != nil {
		return err
	}

	if envelope.Providers.LDAP == nil {
		return errProviderUnconfigured
	}

	a.EnableBasicAuth = envelope.Providers.LDAP.EnableBasicAuth
	a.GroupMembershipFilter = envelope.Providers.LDAP.GroupMembershipFilter
	a.GroupSearchBase = envelope.Providers.LDAP.GroupSearchBase
	a.ManagerDN = envelope.Providers.LDAP.ManagerDN
	a.ManagerPassword = envelope.Providers.LDAP.ManagerPassword
	a.RootDN = envelope.Providers.LDAP.RootDN
	a.Server = envelope.Providers.LDAP.Server
	a.UserSearchBase = envelope.Providers.LDAP.UserSearchBase
	a.UserSearchFilter = envelope.Providers.LDAP.UserSearchFilter
	a.UsernameAttribute = envelope.Providers.LDAP.UsernameAttribute
	a.TLSConfig = envelope.Providers.LDAP.TLSConfig

	// Set defaults
	if a.UserSearchFilter == "" {
		a.UserSearchFilter = `(uid={0})`
	}
	if a.GroupMembershipFilter == "" {
		a.GroupMembershipFilter = `(|(member={0})(uniqueMember={0}))`
	}
	if a.UserSearchBase == "" {
		a.UserSearchBase = a.RootDN
	}

	if a.GroupSearchBase == "" {
		a.GroupSearchBase = a.RootDN
	}

	if a.UsernameAttribute == "" {
		a.UsernameAttribute = "dn"
	}

	return nil
}

// DetectUser is used to detect a user without a login form from
// a cookie, header or other methods
// If no user was detected the errNoValidUserFound needs to be
// returned
func (a authLDAP) DetectUser(res http.ResponseWriter, r *http.Request) (string, []string, error) {
	var alias, user string

	if a.EnableBasicAuth {
		if basicUser, basicPass, ok := r.BasicAuth(); ok {
			if userDN, a, err := a.checkLogin(basicUser, basicPass, a.UsernameAttribute); err != nil {
				return "", nil, err
			} else {
				user = userDN
				alias = a
			}
		}
	}

	if user == "" {
		sess, err := cookieStore.Get(r, strings.Join([]string{mainCfg.Cookie.Prefix, a.AuthenticatorID()}, "-"))
		if err != nil {
			return "", nil, errNoValidUserFound
		}

		var ok bool
		if user, ok = sess.Values["user"].(string); !ok {
			return "", nil, errNoValidUserFound
		}

		if alias, ok = sess.Values["alias"].(string); !ok {
			// Most likely an old cookie, force re-login
			return "", nil, errNoValidUserFound
		}

		// We had a cookie, lets renew it
		sess.Options = mainCfg.GetSessionOpts()
		if err := sess.Save(r, res); err != nil {
			return "", nil, err
		}
	}

	groups, err := a.getUserGroups(user, alias)

	return alias, groups, err
}

// Login is called when the user submits the login form and needs
// to authenticate the user or throw an error. If the user has
// successfully logged in the persistent cookie should be written
// in order to use DetectUser for the next login.
// If the user did not login correctly the errNoValidUserFound
// needs to be returned
func (a authLDAP) Login(res http.ResponseWriter, r *http.Request) (string, []mfaConfig, error) {
	username := r.FormValue(strings.Join([]string{a.AuthenticatorID(), "username"}, "-"))
	password := r.FormValue(strings.Join([]string{a.AuthenticatorID(), "password"}, "-"))

	var (
		userDN string
		alias  string
		err    error
	)

	if userDN, alias, err = a.checkLogin(username, password, a.UsernameAttribute); err != nil {
		return "", nil, err
	}

	sess, _ := cookieStore.Get(r, strings.Join([]string{mainCfg.Cookie.Prefix, a.AuthenticatorID()}, "-"))
	sess.Options = mainCfg.GetSessionOpts()
	sess.Values["user"] = userDN
	sess.Values["alias"] = alias
	return userDN, nil, sess.Save(r, res)
}

// LoginFields needs to return the fields required for this login
// method. If no login using this method is possible the function
// needs to return nil.
func (a authLDAP) LoginFields() (fields []loginField) {
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
func (a authLDAP) Logout(res http.ResponseWriter, r *http.Request) (err error) {
	sess, _ := cookieStore.Get(r, strings.Join([]string{mainCfg.Cookie.Prefix, a.AuthenticatorID()}, "-"))
	sess.Options = mainCfg.GetSessionOpts()
	sess.Options.MaxAge = -1 // Instant delete
	return sess.Save(r, res)
}

// checkLogin searches for the username using the specified UserSearchFilter
// and returns the UserDN and an error (errNoValidUserFound / processing error)
func (a authLDAP) checkLogin(username, password, aliasAttribute string) (string, string, error) {
	l, err := a.dial()
	if err != nil {
		return "", "", err
	}
	defer l.Close()

	sreq := ldap.NewSearchRequest(
		a.UserSearchBase,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0, 0, false,
		strings.Replace(a.UserSearchFilter, `{0}`, username, -1),
		[]string{"dn", aliasAttribute},
		nil,
	)

	sres, err := l.Search(sreq)
	if err != nil {
		return "", "", fmt.Errorf("Unable to search for user: %s", err)
	}

	if len(sres.Entries) != 1 {
		return "", "", errNoValidUserFound
	}

	userDN := sres.Entries[0].DN

	if err := l.Bind(userDN, password); err != nil {
		return "", "", errNoValidUserFound
	}

	alias := sres.Entries[0].GetAttributeValue(aliasAttribute)
	if aliasAttribute == "dn" {
		// DN is not fetchable through GetAttributeValue as it is not an attribute
		alias = userDN
	}

	return userDN, alias, nil
}

func (a authLDAP) portFromScheme(scheme, override string) string {
	if override != "" {
		return override
	}

	switch scheme {
	case "ldap":
		return "389"
	case "ldaps":
		return "636"
	default:
		return ""
	}
}

// dial connects to the LDAP server and authenticates using manager_dn
func (a authLDAP) dial() (*ldap.Conn, error) {
	u, err := url.Parse(a.Server)
	if err != nil {
		return nil, err
	}

	host := u.Host
	port := u.Port()

	var l *ldap.Conn

	switch u.Scheme {
	case "ldap":
		l, err = ldap.Dial("tcp", fmt.Sprintf("%s:%s", host, a.portFromScheme(u.Scheme, port)))

	case "ldaps":
		tlsConfig := &tls.Config{ServerName: host}

		if a.TLSConfig != nil && (a.TLSConfig.ValidateHostname != "" || a.TLSConfig.AllowInsecure) {
			tlsConfig = &tls.Config{
				ServerName:         a.TLSConfig.ValidateHostname,
				InsecureSkipVerify: a.TLSConfig.AllowInsecure,
			}
		}

		l, err = ldap.DialTLS(
			"tcp", fmt.Sprintf("%s:%s", host, a.portFromScheme(u.Scheme, port)),
			tlsConfig,
		)

	default:
		return nil, fmt.Errorf("Unsupported scheme %s", u.Scheme)
	}

	if err != nil {
		return nil, fmt.Errorf("Unable to connect to LDAP: %s", err)
	}

	if err := l.Bind(a.ManagerDN, a.ManagerPassword); err != nil {
		return nil, fmt.Errorf("Unable to authenticate with manager_dn: %s", err)
	}

	return l, err
}

// getUserGroups searches for groups containing the user
func (a authLDAP) getUserGroups(userDN, alias string) ([]string, error) {
	l, err := a.dial()
	if err != nil {
		return nil, err
	}
	defer l.Close()

	sreq := ldap.NewSearchRequest(
		a.GroupSearchBase,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0, 0, false,
		strings.NewReplacer(
			`{0}`, userDN,
			`{1}`, alias,
		).Replace(a.GroupMembershipFilter),
		[]string{"dn"},
		nil,
	)

	sres, err := l.Search(sreq)
	if err != nil {
		return nil, fmt.Errorf("Unable to search for groups: %s", err)
	}

	groups := []string{}
	for _, r := range sres.Entries {
		groups = append(groups, r.DN)
	}

	return groups, nil
}

// SupportsMFA returns the MFA detection capabilities of the login
// provider. If the provider can provide mfaConfig objects from its
// configuration return true. If this is true the login interface
// will display an additional field for this provider for the user
// to fill in their MFA token.
func (a authLDAP) SupportsMFA() bool { return false } // TODO: Implement
