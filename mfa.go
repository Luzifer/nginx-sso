package main

import (
	"fmt"
	"net/http"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/Luzifer/nginx-sso/plugins"
)

const mfaLoginFieldName = "mfa-token"

var mfaLoginField = plugins.LoginField{
	Label:       "MFA Token",
	Name:        mfaLoginFieldName,
	Placeholder: "(optional)",
	Type:        "text",
}

var (
	mfaRegistry      = []plugins.MFAProvider{}
	mfaRegistryMutex sync.RWMutex

	activeMFAProviders = []plugins.MFAProvider{}
)

func registerMFAProvider(m plugins.MFAProvider) {
	mfaRegistryMutex.Lock()
	defer mfaRegistryMutex.Unlock()

	mfaRegistry = append(mfaRegistry, m)
}

func initializeMFAProviders(yamlSource []byte) error {
	mfaRegistryMutex.Lock()
	defer mfaRegistryMutex.Unlock()

	for _, m := range mfaRegistry {
		err := m.Configure(yamlSource)

		switch err {
		case nil:
			activeMFAProviders = append(activeMFAProviders, m)
			log.WithFields(log.Fields{"mfa_provider": m.ProviderID()}).Debug("Activated MFA provider")
		case plugins.ErrProviderUnconfigured:
			log.WithFields(log.Fields{"mfa_provider": m.ProviderID()}).Debug("MFA provider unconfigured")
			// This is okay.
		default:
			return fmt.Errorf("MFA provider configuration caused an error: %s", err)
		}
	}

	return nil
}

func validateMFA(res http.ResponseWriter, r *http.Request, user string, mfaCfgs []plugins.MFAConfig) error {
	if len(mfaCfgs) == 0 {
		// User has no configured MFA devices, their MFA is automatically valid
		return nil
	}

	mfaRegistryMutex.RLock()
	defer mfaRegistryMutex.RUnlock()

	for _, m := range activeMFAProviders {
		err := m.ValidateMFA(res, r, user, mfaCfgs)
		switch err {
		case nil:
			// Validated successfully
			return nil
		case plugins.ErrNoValidUserFound:
			// This is fine for now
		default:
			return err
		}
	}

	// No method could verify the user
	return plugins.ErrNoValidUserFound
}
