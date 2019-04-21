package crowd

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// Session represents a single sign-on (SSO) session in Crowd
type Session struct {
	XMLName struct{}  `xml:"session"`
	Expand  string    `xml:"expand,attr"`
	Token   string    `xml:"token"`
	User    User      `xml:"user,omitempty"`
	Created time.Time `xml:"created-date"`
	Expires time.Time `xml:"expiry-date"`
}

// session authentication request
type sessionAuthReq struct {
	XMLName           struct{}                  `xml:"authentication-context"`
	Username          string                    `xml:"username"`
	Password          string                    `xml:"password"`
	ValidationFactors []sessionValidationFactor `xml:"validation-factors>validation-factor"`
}

// validation factors for session
type sessionValidationFactor struct {
	XMLName struct{} `xml:"validation-factor"`
	Name    string   `xml:"name"`
	Value   string   `xml:"value"`
}

// session validation request -> just validation factors
type sessionValidationValidationFactor struct {
	XMLName           struct{}                  `xml:"validation-factors"`
	ValidationFactors []sessionValidationFactor `xml:"validation-factor"`
}

// NewSession authenticates a user and starts a SSO session if valid.
func (c *Crowd) NewSession(user string, pass string, address string) (Session, error) {
	s := Session{}

	svf := sessionValidationFactor{Name: "remote_address", Value: address}
	sar := sessionAuthReq{Username: user, Password: pass}
	sar.ValidationFactors = append(sar.ValidationFactors, svf)

	sarEncoded, err := xml.Marshal(sar)
	if err != nil {
		return s, err
	}
	sarBuf := bytes.NewBuffer(sarEncoded)

	url := c.url + "rest/usermanagement/1/session"

	c.Client.Jar = c.cookies
	req, err := http.NewRequest("POST", url, sarBuf)
	if err != nil {
		return s, err
	}
	req.SetBasicAuth(c.user, c.passwd)
	req.Header.Set("Accept", "application/xml")
	req.Header.Set("Content-Type", "application/xml")
	resp, err := c.Client.Do(req)
	if err != nil {
		return s, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return s, err
	}

	switch resp.StatusCode {
	case 400:
		er := Error{}
		err = xml.Unmarshal(body, &er)
		if err != nil {
			return s, err
		}

		return s, fmt.Errorf("%s", er.Reason)
	case 201:
		err = xml.Unmarshal(body, &s)
		if err != nil {
			return s, err
		}
	default:
		return s, fmt.Errorf("request failed: %s", resp.Status)
	}

	return s, nil
}

// ValidateSession validates a SSO token against Crowd. Returns error on failure
// or account lockout. Success is a populated Session with nil error.
func (c *Crowd) ValidateSession(token string, clientaddr string) (Session, error) {
	s := Session{}

	svf := sessionValidationFactor{Name: "remote_address", Value: clientaddr}
	svvf := sessionValidationValidationFactor{}
	svvf.ValidationFactors = append(svvf.ValidationFactors, svf)

	svvfEncoded, err := xml.Marshal(svvf)
	if err != nil {
		return s, err
	}
	svvfBuf := bytes.NewBuffer(svvfEncoded)

	url := c.url + "rest/usermanagement/1/session/" + token

	c.Client.Jar = c.cookies
	req, err := http.NewRequest("POST", url, svvfBuf)
	if err != nil {
		return s, err
	}
	req.SetBasicAuth(c.user, c.passwd)
	req.Header.Set("Accept", "application/xml")
	req.Header.Set("Content-Type", "application/xml")
	resp, err := c.Client.Do(req)
	if err != nil {
		return s, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return s, err
	}

	switch resp.StatusCode {
	case 400:
		er := Error{}
		err = xml.Unmarshal(body, &er)
		if err != nil {
			return s, err
		}

		return s, fmt.Errorf("%s", er.Reason)
	case 404:
		er := Error{}
		err = xml.Unmarshal(body, &er)
		if err != nil {
			return s, err
		}

		return s, fmt.Errorf("%s", er.Reason)
	case 200:
		err = xml.Unmarshal(body, &s)
		if err != nil {
			return s, err
		}
	default:
		return s, fmt.Errorf("request failed: %s", resp.Status)
	}

	return s, nil
}

// InvalidateSession invalidates SSO session token. Returns error on failure.
func (c *Crowd) InvalidateSession(token string) error {
	c.Client.Jar = c.cookies
	req, err := http.NewRequest("DELETE", c.url+"rest/usermanagement/1/session/"+token, nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth(c.user, c.passwd)
	req.Header.Set("Accept", "application/xml")
	req.Header.Set("Content-Type", "application/xml")
	resp, err := c.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 204 {
		return fmt.Errorf("request failed: %s", resp.Status)
	}

	return nil
}

// GetSession gets SSO session information by token
func (c *Crowd) GetSession(token string) (s Session, err error) {
	c.Client.Jar = c.cookies
	url := c.url + "rest/usermanagement/1/session/" + token
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return s, err
	}
	req.SetBasicAuth(c.user, c.passwd)
	req.Header.Set("Accept", "application/xml")
	req.Header.Set("Content-Type", "application/xml")
	resp, err := c.Client.Do(req)
	if err != nil {
		return s, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 404:
		return s, fmt.Errorf("session not found")
	case 200:
		// fall through switch without returning
	default:
		return s, fmt.Errorf("request failed: %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return s, err
	}

	err = xml.Unmarshal(body, &s)
	if err != nil {
		return s, err
	}

	return s, nil
}
