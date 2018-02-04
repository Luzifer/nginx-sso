package crowd

import (
	"os"
	"testing"
)

func TestSSOLifeCycle(t *testing.T) {
	tv := PrepVars(t)
	c, err := New(tv.AppUsername, tv.AppPassword, tv.AppURL)
	if err != nil {
		t.Error(err)
	}

	user := os.Getenv("APP_USER_USERNAME")
	if user == "" {
		t.Skip("Can't run test because APP_USER_USERNAME undefined")
	}

	passwd := os.Getenv("APP_USER_PASSWORD")
	if passwd == "" {
		t.Skip("Can't run test because APP_USER_PASSWORD undefined")
	}

	addr := "10.10.10.10"

	// test new session
	a, err := c.NewSession(user, passwd, addr)
	if err != nil {
		t.Errorf("Error creating new session: %s\n", err)
	} else {
		t.Logf("Got new session: %+v\n", a)
	}

	if a.Token == "" {
		t.Errorf("Token was empty so we didn't get/decode a response from NewSession")
	}

	// test validate for session we just created
	si, err := c.ValidateSession(a.Token, addr)
	if err != nil {
		t.Errorf("Error validating session: %s\n", err)
	} else {
		t.Logf("Validated session: %+v\n", si)
	}

	if si.Token == "" {
		t.Errorf("Token was empty so we didn't get/decode a response from ValidateSession")
	}

	// test get info for session
	sdat, err := c.GetSession(a.Token)
	if err != nil {
		t.Errorf("Error getting session: %s\n", err)
	} else {
		t.Logf("Got session: %+v\n", sdat)
	}

	// test invalidating session
	err = c.InvalidateSession(a.Token)
	if err != nil {
		t.Errorf("Error invalidating session: %s\n", err)
	} else {
		t.Log("Invalidated session")
	}

	// make sure sesssion is gone
	ivsess, err := c.ValidateSession(a.Token, addr)
	if err == nil {
		t.Errorf("Validating non-existant session should fail, got: %+v\n", ivsess)
	} else {
		t.Log("Could not validate session that doesn't exist (this is good)")
	}

}
