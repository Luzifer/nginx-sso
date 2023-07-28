package main

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/pkg/errors"
)

func getRedirectURL(r *http.Request, fallback string) (string, error) {
	var (
		params      url.Values
		redirURL    string
		removeParam string
		sessURL     string
	)

	if cookieStore != nil {
		sess, _ := cookieStore.Get(r, strings.Join([]string{mainCfg.Cookie.Prefix, "main"}, "-"))
		if s, ok := sess.Values["go"].(string); ok {
			sessURL = s
		}
	}

	switch {
	case r.URL.Query().Get("rd") != "":
		// We have a GET request with a "rd" query param (K8s ingress-nginx)
		redirURL = r.URL.Query().Get("rd")
		removeParam = "rd"
		params = r.URL.Query()

	case r.URL.Query().Get("go") != "":
		// We have a GET request, use "go" query param
		redirURL = r.URL.Query().Get("go")
		removeParam = "go"
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

	// Remove the redirect parameter as it is a parameter for nginx-sso
	if removeParam != "" {
		params.Del(removeParam)
	}

	// Parse given URL to extract attached parameters
	pURL, err := url.Parse(redirURL)
	if err != nil {
		return "", errors.Wrap(err, "parsing redirect URL")
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
