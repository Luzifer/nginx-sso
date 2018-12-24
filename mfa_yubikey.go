package main

import (
	"net/http"
	"strings"

	"github.com/GeertJohan/yubigo"
	yaml "gopkg.in/yaml.v2"
)

func init() {
	registerMFAProvider(&mfaYubikey{})
}

type mfaYubikey struct {
	ClientID  string `yaml:"client_id"`
	SecretKey string `yaml:"secret_key"`
}

// ProviderID needs to return an unique string to identify
// this special MFA provider
func (m mfaYubikey) ProviderID() (id string) { return "yubikey" }

// Configure loads the configuration for the Authenticator from the
// global config.yaml file which is passed as a byte-slice.
// If no configuration for the Authenticator is supplied the function
// needs to return the errProviderUnconfigured
func (m mfaYubikey) Configure(yamlSource []byte) (err error) {
	envelope := struct {
		MFA struct {
			Yubikey *mfaYubikey `yaml:"yubikey"`
		} `yaml:"mfa"`
	}{}

	if err := yaml.Unmarshal(yamlSource, &envelope); err != nil {
		return err
	}

	if envelope.MFA.Yubikey == nil {
		return errProviderUnconfigured
	}

	m.ClientID = envelope.MFA.Yubikey.ClientID
	m.SecretKey = envelope.MFA.Yubikey.SecretKey

	return nil
}

// ValidateMFA takes the user from the login cookie and performs a
// validation against the provided MFA configuration for this user
func (m mfaYubikey) ValidateMFA(res http.ResponseWriter, r *http.Request, user string, mfaCfgs []mfaConfig) error {
	keyInput := r.FormValue(mfaLoginFieldName)

	yubiAuth, err := yubigo.NewYubiAuth(m.ClientID, m.SecretKey)
	if err != nil {
		return err
	}

	for _, c := range mfaCfgs {
		if c.Provider != m.ProviderID() {
			continue
		}

		if !strings.HasPrefix(keyInput, c.AttributeString("device")) {
			// Might be a valid OTP but is not the key configured in this config
			continue
		}

		_, ok, err := yubiAuth.Verify(keyInput)
		if err != nil && !strings.Contains(err.Error(), "OTP has wrong length.") {
			return err
		}

		if ok {
			return nil
		}
	}

	// Not a valid authentication
	return errNoValidUserFound
}
