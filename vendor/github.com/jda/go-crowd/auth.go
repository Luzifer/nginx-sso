package crowd

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

type authReq struct {
	XMLName  struct{} `xml:"password"`
	Password string   `xml:"value"`
}

// Authenticate a user & password against Crowd. Returns error on failure
// or account lockout. Success is a populated User with nil error.
func (c *Crowd) Authenticate(user string, pass string) (User, error) {
	u := User{}

	ar := authReq{Password: pass}
	arEncoded, err := xml.Marshal(ar)
	if err != nil {
		return u, err
	}
	arBuf := bytes.NewBuffer(arEncoded)

	v := url.Values{}
	v.Set("username", user)
	url := c.url + "rest/usermanagement/1/authentication?" + v.Encode()

	client := http.Client{Jar: c.cookies}
	req, err := http.NewRequest("POST", url, arBuf)
	if err != nil {
		return u, err
	}
	req.SetBasicAuth(c.user, c.passwd)
	req.Header.Set("Accept", "application/xml")
	req.Header.Set("Content-Type", "application/xml")
	resp, err := client.Do(req)
	if err != nil {
		return u, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return u, err
	}

	switch resp.StatusCode {
	case 400:
		er := Error{}
		err = xml.Unmarshal(body, &er)
		if err != nil {
			return u, err
		}

		return u, fmt.Errorf("%s", er.Reason)
	case 200:
		err = xml.Unmarshal(body, &u)
		if err != nil {
			return u, err
		}
	default:
		return u, fmt.Errorf("request failed: %s\n", resp.Status)
	}

	return u, nil
}
