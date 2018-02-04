package crowd

import (
	"os"
	"testing"
)

type TestVars struct {
	AppUsername string
	AppPassword string
	AppURL      string
}

// Make sure we have the env vars to run, handle bailing if we don't
func PrepVars(t *testing.T) TestVars {
	var tv TestVars

	appU := os.Getenv("APP_USERNAME")
	if appU == "" {
		t.Skip("Can't run test because APP_USERNAME undefined")
	} else {
		tv.AppUsername = appU
	}

	appP := os.Getenv("APP_PASSWORD")
	if appP == "" {
		t.Skip("Can't run test because APP_PASSWORD undefined")
	} else {
		tv.AppPassword = appP
	}

	appURL := os.Getenv("APP_URL")
	if appURL == "" {
		t.Skip("Can't run test because APP_URL undefined")
	} else {
		tv.AppURL = appURL
	}

	return tv
}
