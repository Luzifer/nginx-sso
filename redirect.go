package main

import (
	"net/http"
	"net/url"

	"github.com/pkg/errors"
)

func getRedirectURL(r *http.Request, fallback string) (string, error) {
	var (
		redirURL string
		params   url.Values
	)

	switch {
	case r.URL.Query().Get("go") != "":
		// We have a GET request, use "go" query param
		redirURL = r.URL.Query().Get("go")
		params = r.URL.Query()

	case r.FormValue("go") != "":
		// We have a POST request, use "go" form value
		redirURL = r.FormValue("go")
		params = url.Values{} // No need to read other form fields

	default:
		// No URL specified, use specified fallback URL
		return fallback, nil
	}

	// Remove the "go" parameter as it is a parameter for nginx-sso
	params.Del("go")

	// Parse given URL to extract attached parameters
	pURL, err := url.Parse(redirURL)
	if err != nil {
		return "", errors.Wrap(err, "Unable to parse redirect URL")
	}

	// Transfer parameters from URL to params set
	for k, vs := range pURL.Query() {
		for _, v := range vs {
			params.Add(k, v)
		}
	}

	// Re-add assembled parameters to URL
	pURL.RawQuery = params.Encode()

	return pURL.String(), nil
}
