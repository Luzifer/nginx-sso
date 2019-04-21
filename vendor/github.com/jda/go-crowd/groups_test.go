package crowd

import (
	"os"
	"testing"
)

func TestGetDirectGroups(t *testing.T) {
	tv := PrepVars(t)
	c, err := New(tv.AppUsername, tv.AppPassword, tv.AppURL)
	if err != nil {
		t.Error(err)
	}

	user := os.Getenv("APP_USER_USERNAME")
	if user == "" {
		t.Skip("Can't run test because APP_USER_USERNAME undefined")
	}

	// test new session
	groups, err := c.GetDirectGroups(user)
	if err != nil {
		t.Errorf("Error getting user's direct group membership list: %s\n", err)
	} else {
		t.Logf("Got user's direct group membership list:")
		for _, element := range groups {
			t.Logf(" %s", element.Name)
		}
	}

	if len(groups) == 0 {
		t.Error("groups list was empty so we didn't get/decode a response from GetIndirectGroups")
	}
}

func TestGetNestedGroups(t *testing.T) {
	tv := PrepVars(t)
	c, err := New(tv.AppUsername, tv.AppPassword, tv.AppURL)
	if err != nil {
		t.Error(err)
	}

	user := os.Getenv("APP_USER_USERNAME")
	if user == "" {
		t.Skip("Can't run test because APP_USER_USERNAME undefined")
	}

	// test new session
	groups, err := c.GetNestedGroups(user)
	if err != nil {
		t.Errorf("Error getting user's nested group membership list: %s\n", err)
	} else {
		t.Logf("Got user's nested group membership list:")
		for _, element := range groups {
			t.Logf(" %s", element.Name)
		}
	}

	if len(groups) == 0 {
		t.Error("groups list was empty so we didn't get/decode a response from GetIndirectGroups")
	}
}

func TestCreateGroup(t *testing.T) {
	tv := PrepVars(t)
	c, err := New(tv.AppUsername, tv.AppPassword, tv.AppURL)
	if err != nil {
		t.Error(err)
	}

	user := os.Getenv("APP_USER_USERNAME")
	if user == "" {
		t.Skip("Can't run test because APP_USER_USERNAME undefined")
	}

	status := c.CreateGroup("test", "test group")
	if !status {
		t.Error("Expected a group to be created")
	}
}

func TestGetGroup(t *testing.T) {
	tv := PrepVars(t)
	c, err := New(tv.AppUsername, tv.AppPassword, tv.AppURL)
	if err != nil {
		t.Error(err)
	}

	user := os.Getenv("APP_USER_USERNAME")
	if user == "" {
		t.Skip("Can't run test because APP_USER_USERNAME undefined")
	}

	group, err := c.GetGroup("test")
	if group == nil {
		t.Error("Expected group attributes.")
	}
	if err != nil {
		t.Errorf("Error given: %s", err)
	}

}
