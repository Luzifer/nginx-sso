package main

import (
	"net/http"
	"net/url"

	"github.com/pkg/errors"
)

func getRedirectURL(r *http.Request) (string, error) {
	var (
		redirURL = r.URL.Query().Get("go")
		params   = r.URL.Query()
	)

	if redirURL == "" {
		redirURL = r.FormValue("go")
		params = url.Values{} // No need to read other form fields
	}

	params.Del("go")

	pURL, err := url.Parse(redirURL)
	if err != nil {
		return "", errors.Wrap(err, "Unable to parse redirect URL")
	}

	for k, vs := range pURL.Query() {
		for _, v := range vs {
			params.Add(k, v)
		}
	}

	pURL.RawQuery = params.Encode()

	return pURL.String(), nil
}
