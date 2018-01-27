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
	// global config.hcl file which is passed as a byte-slice.
	// If no configuration for the Authenticator is supplied the function
	// needs to return the authenticatorUnconfiguredError
	Configure(hclSource []byte) (err error)

	// DetectUser is used to detect a user without a login form from
	// a cookie, header or other methods
	// If no user was detected the noValidUserFoundError needs to be
	// returned
	DetectUser(r *http.Request) (user string, groups []string, err error)

	// Login is called when the user submits the login form and needs
	// to authenticate the user or throw an error. If the user has
	// successfully logged in the persistent cookie should be written
	// in order to use DetectUser for the next login.
	// If the user did not login correctly the noValidUserFoundError
	// needs to be returned
	Login(res http.ResponseWriter, r *http.Request) (err error)

	// LoginFields needs to return the fields required for this login
	// method. If no login using this method is possible the function
	// needs to return nil.
	LoginFields() (fields []loginField)

	// Logout is called when the user visits the logout endpoint and
	// needs to destroy any persistent stored cookies
	Logout(res http.ResponseWriter) (err error)
}

type loginField struct {
	Label       string
	Name        string
	Placeholder string
	Type        string
}

var (
	authenticatorUnconfiguredError = errors.New("No valid configuration found for this authenticator")
	noValidUserFoundError          = errors.New("No valid users found")

	authenticatorRegistry      = []authenticator{}
	authenticatorRegistryMutex sync.RWMutex

	activeAuthenticators = []authenticator{}
)

func registerAuthenticator(a authenticator) {
	authenticatorRegistryMutex.Lock()
	defer authenticatorRegistryMutex.Unlock()

	authenticatorRegistry = append(authenticatorRegistry, a)
}

func initializeAuthenticators(hclSource []byte) error {
	authenticatorRegistryMutex.Lock()
	defer authenticatorRegistryMutex.Unlock()

	for _, a := range authenticatorRegistry {
		err := a.Configure(hclSource)

		switch err {
		case nil:
			activeAuthenticators = append(activeAuthenticators, a)
			log.WithFields(log.Fields{"authenticator": a.AuthenticatorID()}).Debug("Activated authenticator")
		case authenticatorUnconfiguredError:
			log.WithFields(log.Fields{"authenticator": a.AuthenticatorID()}).Debug("Authenticator unconfigured")
			// This is okay.
		default:
			return fmt.Errorf("Authenticator configuration caused an error: %s", err)
		}
	}

	if len(activeAuthenticators) == 0 {
		return fmt.Errorf("No authenticator configurations supplied")
	}

	return nil
}

func detectUser(r *http.Request) (string, []string, error) {
	authenticatorRegistryMutex.RLock()
	defer authenticatorRegistryMutex.RUnlock()

	for _, a := range activeAuthenticators {
		user, groups, err := a.DetectUser(r)
		switch err {
		case nil:
			return user, groups, err
		case noValidUserFoundError:
			// This is okay.
		default:
			return "", nil, err
		}
	}

	return "", nil, noValidUserFoundError
}

func loginUser(res http.ResponseWriter, r *http.Request) error {
	authenticatorRegistryMutex.RLock()
	defer authenticatorRegistryMutex.RUnlock()

	for _, a := range activeAuthenticators {
		err := a.Login(res, r)
		switch err {
		case nil:
			return nil
		case noValidUserFoundError:
			// This is okay.
		default:
			return err
		}
	}

	return noValidUserFoundError
}
