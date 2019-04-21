package crowd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

// Link is a child of Group
type Link struct {
	Href string `json:"href"`
	Rel  string `json:"rel"`
}

// Group represents a group in Crowd
type Group struct {
	Name string `json:"name"`
	Link Link   `json:"link"`
}

// Groups come in lists
type listGroups struct {
	Groups []*Group `json:"groups"`
	Expand string   `json:"expand"`
}

// GroupInfo returns information about a group
type GroupInfo struct {
	Expand string `json:"expand"`
	Link   struct {
		Href string `json:"href"`
		Rel  string `json:"rel"`
	} `json:"link"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Active      bool   `json:"active"`
	Attributes  struct {
		Attributes []interface{} `json:"attributes"`
		Link       struct {
			Href string `json:"href"`
			Rel  string `json:"rel"`
		} `json:"link"`
	} `json:"attributes"`
}

// GetGroups retrieves a list of groups of which a user is a direct (and nested if donested is true) member.
func (c *Crowd) GetGroups(user string, donested bool) ([]*Group, error) {
	groupList := listGroups{}

	v := url.Values{}
	v.Set("username", user)
	var endpoint string

	if donested {
		endpoint = "nested"
	} else {
		endpoint = "direct"
	}

	url := c.url + "rest/usermanagement/1/user/group/" + endpoint + "?" + v.Encode()
	c.Client.Jar = c.cookies
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return groupList.Groups, err
	}
	req.SetBasicAuth(c.user, c.passwd)
	req.Header.Set("Accept", "application/json")
	resp, err := c.Client.Do(req)
	if err != nil {
		return groupList.Groups, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 404:
		return groupList.Groups, fmt.Errorf("user not found")
	case 200:
		// fall through switch without returning
	default:
		return groupList.Groups, fmt.Errorf("request failed: %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return groupList.Groups, err
	}

	err = json.Unmarshal(body, &groupList)
	if err != nil {
		return groupList.Groups, err
	}

	return groupList.Groups, nil
}

// GetNestedGroups retrieves a list of groups of which a user is a direct or nested member
func (c *Crowd) GetNestedGroups(user string) ([]*Group, error) {
	return c.GetGroups(user, true)
}

// GetDirectGroups retrieves a list of groups of which a user is a direct member
func (c *Crowd) GetDirectGroups(user string) ([]*Group, error) {
	return c.GetGroups(user, false)
}

// GetGroup returns a group
func (c *Crowd) GetGroup(name string) (*Group, error) {
	attrURL := fmt.Sprintf("rest/usermanagement/1/group?groupname=%s&expand=attributes", name)
	url := c.url + attrURL

	c.Client.Jar = c.cookies

	req, err := http.NewRequest("GET", url, nil)
	req.SetBasicAuth(c.user, c.passwd)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if err != nil {
		panic(err)
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()
	groupInformation, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	groupAttributes := new(Group)
	err = json.Unmarshal(groupInformation, groupAttributes)
	if err != nil {
		panic(err)
	}

	return groupAttributes, err
}

// CreateGroup creates a new group
func (c *Crowd) CreateGroup(name string, description string) (status bool) {

	url := c.url + "rest/usermanagement/1/group"

	c.Client.Jar = c.cookies

	values := map[string]string{"name": name, "type": "GROUP", "description": description, "active": "true"}
	jsonStr, _ := json.Marshal(values)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		panic(err)
	}
	req.SetBasicAuth(c.user, c.passwd)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.Client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 201:
		status = true
	default:
		fmt.Printf("request failed: %s\n", resp.Status)
	}

	return status

}
