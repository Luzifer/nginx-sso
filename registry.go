package main

import (
	"errors"
	"fmt"
	"net/http"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/Luzifer/nginx-sso/plugins"
)

var (
	errProviderUnconfigured = errors.New("No valid configuration found for this provider")
	errNoValidUserFound     = errors.New("No valid users found")

	authenticatorRegistry      = []plugins.Authenticator{}
	authenticatorRegistryMutex sync.RWMutex

	activeAuthenticators = []plugins.Authenticator{}
)

func registerAuthenticator(a plugins.Authenticator) {
	authenticatorRegistryMutex.Lock()
	defer authenticatorRegistryMutex.Unlock()

	authenticatorRegistry = append(authenticatorRegistry, a)
}

func initializeAuthenticators(yamlSource []byte) error {
	authenticatorRegistryMutex.Lock()
	defer authenticatorRegistryMutex.Unlock()

	tmp := []plugins.Authenticator{}
	for _, a := range authenticatorRegistry {
		err := a.Configure(yamlSource)

		switch err {
		case nil:
			tmp = append(tmp, a)
			log.WithFields(log.Fields{"authenticator": a.AuthenticatorID()}).Debug("Activated authenticator")
		case errProviderUnconfigured:
			log.WithFields(log.Fields{"authenticator": a.AuthenticatorID()}).Debug("Authenticator unconfigured")
			// This is okay.
		default:
			return fmt.Errorf("Authenticator configuration caused an error: %s", err)
		}
	}

	if len(tmp) == 0 {
		return fmt.Errorf("No authenticator configurations supplied")
	}

	activeAuthenticators = tmp

	return nil
}

func detectUser(res http.ResponseWriter, r *http.Request) (string, []string, error) {
	authenticatorRegistryMutex.RLock()
	defer authenticatorRegistryMutex.RUnlock()

	for _, a := range activeAuthenticators {
		user, groups, err := a.DetectUser(res, r)
		switch err {
		case nil:
			return user, groups, err
		case errNoValidUserFound:
			// This is okay.
		default:
			return "", nil, err
		}
	}

	return "", nil, errNoValidUserFound
}

func loginUser(res http.ResponseWriter, r *http.Request) (string, []plugins.MFAConfig, error) {
	authenticatorRegistryMutex.RLock()
	defer authenticatorRegistryMutex.RUnlock()

	for _, a := range activeAuthenticators {
		user, mfaCfgs, err := a.Login(res, r)
		switch err {
		case nil:
			return user, mfaCfgs, nil
		case errNoValidUserFound:
			// This is okay.
		default:
			return "", nil, err
		}
	}

	return "", nil, errNoValidUserFound
}

func logoutUser(res http.ResponseWriter, r *http.Request) error {
	authenticatorRegistryMutex.RLock()
	defer authenticatorRegistryMutex.RUnlock()

	for _, a := range activeAuthenticators {
		if err := a.Logout(res, r); err != nil {
			return err
		}
	}

	return nil
}

func getFrontendAuthenticators() map[string][]plugins.LoginField {
	authenticatorRegistryMutex.RLock()
	defer authenticatorRegistryMutex.RUnlock()

	output := map[string][]plugins.LoginField{}
	for _, a := range activeAuthenticators {
		if len(a.LoginFields()) == 0 {
			continue
		}
		output[a.AuthenticatorID()] = a.LoginFields()

		if a.SupportsMFA() && !mainCfg.Login.HideMFAField {
			output[a.AuthenticatorID()] = append(output[a.AuthenticatorID()], mfaLoginField)
		}
	}

	return output
}
