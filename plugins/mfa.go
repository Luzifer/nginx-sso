package plugins

import "net/http"

const MFALoginFieldName = "mfa-token"

type MFAProvider interface {
	// ProviderID needs to return an unique string to identify
	// this special MFA provider
	ProviderID() (id string)

	// Configure loads the configuration for the Authenticator from the
	// global config.yaml file which is passed as a byte-slice.
	// If no configuration for the Authenticator is supplied the function
	// needs to return the ErrProviderUnconfigured
	Configure(yamlSource []byte) (err error)

	// ValidateMFA takes the user from the login cookie and performs a
	// validation against the provided MFA configuration for this user
	ValidateMFA(res http.ResponseWriter, r *http.Request, user string, mfaCfgs []MFAConfig) error
}

type MFAConfig struct {
	Provider   string                 `yaml:"provider"`
	Attributes map[string]interface{} `yaml:"attributes"`
}

func (m MFAConfig) AttributeInt(key string) int {
	if v, ok := m.Attributes[key]; ok && v != "" {
		if sv, ok := v.(int); ok {
			return sv
		}
	}

	return 0
}

func (m MFAConfig) AttributeString(key string) string {
	if v, ok := m.Attributes[key]; ok {
		if sv, ok := v.(string); ok {
			return sv
		}
	}

	return ""
}
