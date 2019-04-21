package crowd

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

// User represents a user in Crowd
type User struct {
	XMLName     struct{} `xml:"user"`
	UserName    string   `xml:"name,attr"`
	FirstName   string   `xml:"first-name"`
	LastName    string   `xml:"last-name"`
	DisplayName string   `xml:"display-name"`
	Email       string   `xml:"email"`
	Active      bool     `xml:"active"`
	Key         string   `xml:"key"`
}

// GetUser retrieves user information
func (c *Crowd) GetUser(user string) (User, error) {
	u := User{}

	v := url.Values{}
	v.Set("username", user)
	url := c.url + "rest/usermanagement/1/user?" + v.Encode()
	c.Client.Jar = c.cookies
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return u, err
	}
	req.SetBasicAuth(c.user, c.passwd)
	req.Header.Set("Accept", "application/xml")
	req.Header.Set("Content-Type", "application/xml")
	resp, err := c.Client.Do(req)
	if err != nil {
		return u, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 404:
		return u, fmt.Errorf("user not found")
	case 200:
		// fall through switch without returning
	default:
		return u, fmt.Errorf("request failed: %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return u, err
	}

	err = xml.Unmarshal(body, &u)
	if err != nil {
		return u, err
	}

	return u, nil
}
