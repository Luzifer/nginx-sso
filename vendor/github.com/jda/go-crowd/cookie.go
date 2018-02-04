package crowd

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
)

// CookieConfig holds configuration values needed to set a Crowd SSO cookie.
type CookieConfig struct {
	XMLName struct{} `xml:"cookie-config"`
	Domain  string   `xml:"domain"`
	Secure  bool     `xml:"secure"`
	Name    string   `xml:"name"`
}

// GetCookieConfig returns settings needed to set a Crowd SSO cookie.
func (c *Crowd) GetCookieConfig() (CookieConfig, error) {
	cc := CookieConfig{}

	client := http.Client{Jar: c.cookies}
	req, err := http.NewRequest("GET", c.url+"rest/usermanagement/1/config/cookie", nil)
	if err != nil {
		return cc, err
	}
	req.SetBasicAuth(c.user, c.passwd)
	req.Header.Set("Accept", "application/xml")
	req.Header.Set("Content-Type", "application/xml")
	resp, err := client.Do(req)
	if err != nil {
		return cc, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return cc, fmt.Errorf("request failed: %s\n", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return cc, err
	}

	err = xml.Unmarshal(body, &cc)
	if err != nil {
		return cc, err
	}

	return cc, nil
}
