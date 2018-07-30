package main

import (
	"fmt"
	"net/http"
	"sync"

	log "github.com/sirupsen/logrus"
)

type mfaProvider interface {
	// ProviderID needs to return an unique string to identify
	// this special MFA provider
	ProviderID() (id string)

	// Configure loads the configuration for the Authenticator from the
	// global config.yaml file which is passed as a byte-slice.
	// If no configuration for the Authenticator is supplied the function
	// needs to return the errProviderUnconfigured
	Configure(yamlSource []byte) (err error)

	// ValidateMFA takes the user from the login cookie and performs a
	// validation against the provided MFA configuration for this user
	ValidateMFA(res http.ResponseWriter, r *http.Request, user string) error

	// UserHasMFA returns true in case there is a MFA configuration for
	// the provided user within this MFA provider
	UserHasMFA(user string) bool
}

var (
	mfaRegistry      = []mfaProvider{}
	mfaRegistryMutex sync.RWMutex

	activeMFAProviders = []mfaProvider{}
)

func registerMFAProvider(m mfaProvider) {
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
		case errProviderUnconfigured:
			log.WithFields(log.Fields{"mfa_provider": m.ProviderID()}).Debug("MFA provider unconfigured")
			// This is okay.
		default:
			return fmt.Errorf("MFA provider configuration caused an error: %s", err)
		}
	}

	return nil
}

func userHasMFA(user string) bool {
	mfaRegistryMutex.RLock()
	defer mfaRegistryMutex.RUnlock()

	for _, m := range activeMFAProviders {
		if m.UserHasMFA(user) {
			return true
		}
	}

	return false
}

func validateMFA(res http.ResponseWriter, r *http.Request) error {
	user, _, err := detectUser(res, r)
	if err != nil {
		return err
	}

	if !userHasMFA(user) {
		// User has no configured MFA devices, their MFA is automatically valid
		return nil
	}

	mfaRegistryMutex.RLock()
	defer mfaRegistryMutex.RUnlock()

	for _, m := range activeMFAProviders {
		err = m.ValidateMFA(res, r, user)
		switch err {
		case nil:
			// Validated successfully
			return nil
		case errNoValidUserFound:
			// This is fine for now
		default:
			return err
		}
	}

	// No method could verify the user
	return errNoValidUserFound
}
