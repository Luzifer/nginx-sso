package crowd

import (
	"testing"
)

func TestGetCookieConfig(t *testing.T) {
	tv := PrepVars(t)
	c, err := New(tv.AppUsername, tv.AppPassword, tv.AppURL)
	if err != nil {
		t.Error(err)
	}

	ck, err := c.GetCookieConfig()
	if err != nil {
		t.Error(err)
	}
	t.Logf("Got: %+v\n", ck)
}
