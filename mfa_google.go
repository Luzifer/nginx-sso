package main

import (
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/pquerna/otp/totp"
)

func init() {
	registerMFAProvider(&mfaGoogle{})
}

type mfaGoogle struct{}

// ProviderID needs to return an unique string to identify
// this special MFA provider
func (m mfaGoogle) ProviderID() (id string) {
	return "google"
}

// Configure loads the configuration for the Authenticator from the
// global config.yaml file which is passed as a byte-slice.
// If no configuration for the Authenticator is supplied the function
// needs to return the errProviderUnconfigured
func (m mfaGoogle) Configure(yamlSource []byte) (err error) { return nil }

// ValidateMFA takes the user from the login cookie and performs a
// validation against the provided MFA configuration for this user
func (m mfaGoogle) ValidateMFA(res http.ResponseWriter, r *http.Request, user string, mfaCfgs []mfaConfig) error {
	// Look for mfaConfigs with own provider name
	for _, c := range mfaCfgs {
		if c.Provider != m.ProviderID() {
			continue
		}

		token, err := m.exec(c)
		if err != nil {
			return errors.Wrap(err, "Generating the MFA token failed")
		}

		if r.FormValue(mfaLoginFieldName) == token {
			return nil
		}
	}

	// Report this provider was not able to verify the MFA request
	return errNoValidUserFound
}

func (m mfaGoogle) exec(c mfaConfig) (string, error) {
	secret := c.AttributeString("secret")

	if n := len(secret) % 8; n != 0 {
		secret = secret + strings.Repeat("=", 8-n)
	}

	return totp.GenerateCode(strings.ToUpper(secret), time.Now())
}
