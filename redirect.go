package main

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/pkg/errors"
)

func getRedirectURL(r *http.Request, fallback string) (string, error) {
	var (
		params   url.Values
		redirURL string
		sessURL  string
	)

	if cookieStore != nil {
		sess, _ := cookieStore.Get(r, strings.Join([]string{mainCfg.Cookie.Prefix, "main"}, "-"))
		if s, ok := sess.Values["go"].(string); ok {
			sessURL = s
		}
	}

	switch {
	case r.URL.Query().Get("go") != "":
		// We have a GET request, use "go" query param
		redirURL = r.URL.Query().Get("go")
		params = r.URL.Query()

	case r.FormValue("go") != "":
		// We have a POST request, use "go" form value
		redirURL = r.FormValue("go")
		params = url.Values{} // No need to read other form fields

	case sessURL != "":
		redirURL = sessURL
		params = url.Values{}

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
