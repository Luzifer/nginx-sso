package main

import (
	"errors"
	"fmt"
	"net/http"
	"sync"

	log "github.com/sirupsen/logrus"
)

type authenticator interface {
	// AuthenticatorID needs to return an unique string to identify
	// this special authenticator
	AuthenticatorID() (id string)

	// Configure loads the configuration for the Authenticator from the
	// global config.yaml file which is passed as a byte-slice.
	// If no configuration for the Authenticator is supplied the function
	// needs to return the errAuthenticatorUnconfigured
	Configure(yamlSource []byte) (err error)

	// DetectUser is used to detect a user without a login form from
	// a cookie, header or other methods
	// If no user was detected the errNoValidUserFound needs to be
	// returned
	DetectUser(res http.ResponseWriter, r *http.Request) (user string, groups []string, err error)

	// Login is called when the user submits the login form and needs
	// to authenticate the user or throw an error. If the user has
	// successfully logged in the persistent cookie should be written
	// in order to use DetectUser for the next login.
	// If the user did not login correctly the errNoValidUserFound
	// needs to be returned
	Login(res http.ResponseWriter, r *http.Request) (err error)

	// LoginFields needs to return the fields required for this login
	// method. If no login using this method is possible the function
	// needs to return nil.
	LoginFields() (fields []loginField)

	// Logout is called when the user visits the logout endpoint and
	// needs to destroy any persistent stored cookies
	Logout(res http.ResponseWriter, r *http.Request) (err error)
}

type loginField struct {
	Label       string
	Name        string
	Placeholder string
	Type        string
}

var (
	errAuthenticatorUnconfigured = errors.New("No valid configuration found for this authenticator")
	errNoValidUserFound          = errors.New("No valid users found")

	authenticatorRegistry      = []authenticator{}
	authenticatorRegistryMutex sync.RWMutex

	activeAuthenticators = []authenticator{}
)

func registerAuthenticator(a authenticator) {
	authenticatorRegistryMutex.Lock()
	defer authenticatorRegistryMutex.Unlock()

	authenticatorRegistry = append(authenticatorRegistry, a)
}

func initializeAuthenticators(yamlSource []byte) error {
	authenticatorRegistryMutex.Lock()
	defer authenticatorRegistryMutex.Unlock()

	tmp := []authenticator{}
	for _, a := range authenticatorRegistry {
		err := a.Configure(yamlSource)

		switch err {
		case nil:
			tmp = append(tmp, a)
			log.WithFields(log.Fields{"authenticator": a.AuthenticatorID()}).Debug("Activated authenticator")
		case errAuthenticatorUnconfigured:
			log.WithFields(log.Fields{"authenticator": a.AuthenticatorID()}).Debug("Authenticator unconfigured")
			// This is okay.
		default:
			return fmt.Errorf("Authenticator configuration caused an error: %s", err)
		}
	}

	if len(activeAuthenticators) == 0 {
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

func loginUser(res http.ResponseWriter, r *http.Request) error {
	authenticatorRegistryMutex.RLock()
	defer authenticatorRegistryMutex.RUnlock()

	for _, a := range activeAuthenticators {
		err := a.Login(res, r)
		switch err {
		case nil:
			return nil
		case errNoValidUserFound:
			// This is okay.
		default:
			return err
		}
	}

	return errNoValidUserFound
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

func getFrontendAuthenticators() map[string][]loginField {
	authenticatorRegistryMutex.RLock()
	defer authenticatorRegistryMutex.RUnlock()

	output := map[string][]loginField{}
	for _, a := range activeAuthenticators {
		if len(a.LoginFields()) == 0 {
			continue
		}
		output[a.AuthenticatorID()] = a.LoginFields()
	}

	return output
}
