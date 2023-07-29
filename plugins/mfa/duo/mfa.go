package duo

import (
	"net"
	"net/http"
	"strings"
	"time"

	duoapi "github.com/duosecurity/duo_api_golang"
	"github.com/duosecurity/duo_api_golang/authapi"
	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v3"

	"github.com/Luzifer/nginx-sso/plugins"
)

const (
	mfaDuoResponseAllow  = "allow"
	mfaDuoRequestTimeout = 10 * time.Second
)

var mfaDuoTrustedIPHeaders = []string{"X-Forwarded-For", "X-Real-IP"}

type MFADuo struct {
	IKey      string `yaml:"ikey"`
	SKey      string `yaml:"skey"`
	Host      string `yaml:"host"`
	UserAgent string `yaml:"user_agent"`
}

func New() *MFADuo {
	return &MFADuo{}
}

// ProviderID needs to return an unique string to identify
// this special MFA provider
func (m MFADuo) ProviderID() (id string) { return "duo" }

// Configure loads the configuration for the Authenticator from the
// global config.yaml file which is passed as a byte-slice.
// If no configuration for the Authenticator is supplied the function
// needs to return the plugins.ErrProviderUnconfigured
func (m *MFADuo) Configure(yamlSource []byte) (err error) {
	envelope := struct {
		MFA struct {
			Duo *MFADuo `yaml:"duo"`
		} `yaml:"mfa"`
	}{}

	if err := yaml.Unmarshal(yamlSource, &envelope); err != nil {
		return err
	}

	if envelope.MFA.Duo == nil {
		return plugins.ErrProviderUnconfigured
	}

	m.IKey = envelope.MFA.Duo.IKey
	m.SKey = envelope.MFA.Duo.SKey
	m.Host = envelope.MFA.Duo.Host
	m.UserAgent = envelope.MFA.Duo.UserAgent
	return nil
}

// ValidateMFA takes the user from the login cookie and performs a
// validation against the provided MFA configuration for this user
func (m MFADuo) ValidateMFA(res http.ResponseWriter, r *http.Request, user string, mfaCfgs []plugins.MFAConfig) error {
	var keyInput string

	// Look for mfaConfigs with own provider name
	for _, c := range mfaCfgs {
		if c.Provider != m.ProviderID() {
			continue
		}

		remoteIP, err := m.findIP(r)
		if err != nil {
			return errors.Wrap(err, "Unable to determine remote IP")
		}

		duo := authapi.NewAuthApi(*duoapi.NewDuoApi(m.IKey, m.SKey, m.Host, m.UserAgent, duoapi.SetTimeout(mfaDuoRequestTimeout)))

		for key, values := range r.Form {
			if strings.HasSuffix(key, plugins.MFALoginFieldName) && len(values[0]) > 0 {
				keyInput = values[0]
			}
		}

		// Check if MFA token provided and fallover to push if not supplied
		var auth *authapi.AuthResult

		if keyInput != "" {
			if auth, err = duo.Auth("passcode", authapi.AuthUsername(user), authapi.AuthPasscode(keyInput), authapi.AuthIpAddr(remoteIP)); err != nil {
				return errors.Wrap(err, "Unable to authenticate with Duo using 'passcode' method")
			}
		} else {
			if auth, err = duo.Auth("auto", authapi.AuthUsername(user), authapi.AuthDevice("auto"), authapi.AuthIpAddr(remoteIP)); err != nil {
				return errors.Wrap(err, "Unable to authenticate with Duo using 'auto' method")
			}
		}

		if auth.Response.Result == mfaDuoResponseAllow {
			return nil
		}
	}

	// Report this provider was not able to verify the MFA request
	return plugins.ErrNoValidUserFound
}

func (m MFADuo) findIP(r *http.Request) (string, error) {
	for _, hdr := range mfaDuoTrustedIPHeaders {
		if value := r.Header.Get(hdr); value != "" {
			return m.parseIP(value)
		}
	}

	return m.parseIP(r.RemoteAddr)
}

func (m MFADuo) parseIP(s string) (string, error) {
	ip, _, err := net.SplitHostPort(s)
	if err == nil {
		return ip, nil
	}

	ip2 := net.ParseIP(s)
	if ip2 == nil {
		return "", errors.New("invalid IP")
	}

	return ip2.String(), nil
}
