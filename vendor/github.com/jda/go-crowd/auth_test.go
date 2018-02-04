package crowd

import (
	"os"
	"testing"
)

func TestAuthentication(t *testing.T) {
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

	a, err := c.Authenticate(user, passwd)
	if err != nil {
		t.Error(err)
	}
	t.Logf("Got: %+v\n", a)

	if a.UserName == "" {
		t.Errorf("UserName was empty so we didn't get/decode a response")
	}
}
