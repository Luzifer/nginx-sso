package main

import (
	"net/http"
	"strings"
	"time"

	"github.com/duosecurity/duo_api_golang"
	"github.com/duosecurity/duo_api_golang/authapi"
	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
)

const mfaDuoResponseAllow = "allow"
const mfaDuoRequestTimeout = 10 * time.Second

func init() {
	registerMFAProvider(&mfaDuo{})
}

type mfaDuo struct {
	IKey      string `yaml:"ikey"`
	SKey      string `yaml:"skey"`
	Host      string `yaml:"host"`
	UserAgent string `yaml:"user_agent"`
}

// ProviderID needs to return an unique string to identify
// this special MFA provider
func (m mfaDuo) ProviderID() (id string) { return "duo" }

// Configure loads the configuration for the Authenticator from the
// global config.yaml file which is passed as a byte-slice.
// If no configuration for the Authenticator is supplied the function
// needs to return the errProviderUnconfigured
func (m *mfaDuo) Configure(yamlSource []byte) (err error) {
	envelope := struct {
		MFA struct {
			Duo *mfaDuo `yaml:"duo"`
		} `yaml:"mfa"`
	}{}

	if err := yaml.Unmarshal(yamlSource, &envelope); err != nil {
		return err
	}

	if envelope.MFA.Duo == nil {
		return errProviderUnconfigured
	}

	m.IKey = envelope.MFA.Duo.IKey
	m.SKey = envelope.MFA.Duo.SKey
	m.Host = envelope.MFA.Duo.Host
	m.UserAgent = envelope.MFA.Duo.UserAgent
	return nil
}

// ValidateMFA takes the user from the login cookie and performs a
// validation against the provided MFA configuration for this user
func (m mfaDuo) ValidateMFA(res http.ResponseWriter, r *http.Request, user string, mfaCfgs []mfaConfig) error {
	var keyInput string
	// Look for mfaConfigs with own provider name
	for _, c := range mfaCfgs {
		if c.Provider != m.ProviderID() {
			continue
		}
		remoteIP := r.Header.Get("X-Real-IP")
		duo := authapi.NewAuthApi(*duoapi.NewDuoApi(m.IKey, m.SKey, m.Host, m.UserAgent, duoapi.SetTimeout((mfaDuoRequestTimeout))))
		for key, values := range r.Form {
			if strings.HasSuffix(key, mfaLoginFieldName) && len(values[0]) > 0 {
				keyInput = values[0]
			}
		}
		//Check if MFA token provided and fallover to push if not supplied
		if keyInput != "" {
			auth, err := duo.Auth("passcode", authapi.AuthUsername(user), authapi.AuthPasscode(keyInput), authapi.AuthIpAddr(remoteIP))
			if err != nil {
				return errors.Wrap(err, "Unable to authenticate with Duo.")
			}
			if auth.Response.Result == mfaDuoResponseAllow {
				return nil
			}
		} else {
			auth, err := duo.Auth("auto", authapi.AuthUsername(user), authapi.AuthDevice("auto"), authapi.AuthIpAddr(remoteIP))
			if err != nil {
				return errors.Wrap(err, "Unable to authenticate with Duo.")
			}
			if auth.Response.Result == mfaDuoResponseAllow {
				return nil
			}
		}
	}

	// Report this provider was not able to verify the MFA request
	return errNoValidUserFound
}
