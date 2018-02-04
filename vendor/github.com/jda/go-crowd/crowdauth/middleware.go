// Package crowdauth provides middleware for Crowd SSO logins
//
// Goals:
//  1) drop-in authentication against Crowd SSO
//  2) make it easy to use Crowd SSO as part of your own auth flow
package crowdauth // import "go.jona.me/crowd/crowdauth"

import (
	"errors"
	"go.jona.me/crowd"
	"html/template"
	"log"
	"net/http"
	"time"
)

type SSO struct {
	CrowdApp            *crowd.Crowd
	LoginPage           AuthLoginPage
	LoginTemplate       *template.Template
	ClientAddressFinder ClientAddressFinder
	CookieConfig        crowd.CookieConfig
}

// The AuthLoginPage type extends the normal http.HandlerFunc type
// with a boolean return to indicate login success or failure.
type AuthLoginPage func(http.ResponseWriter, *http.Request, *SSO) bool

// ClientAddressFinder type represents a function that returns the
// end-user's IP address, allowing you to handle cases where the address
// is masked by proxy servers etc by checking headers or whatever to find
// the user's address
type ClientAddressFinder func(*http.Request) (string, error)

var authErr string = "unauthorized, login required"

func DefaultClientAddressFinder(r *http.Request) (addr string, err error) {
	return r.RemoteAddr, nil
}

// New creates a new instance of SSO
func New(user string, password string, crowdURL string) (s *SSO, err error) {
	s = &SSO{}
	s.LoginPage = loginPage
	s.ClientAddressFinder = DefaultClientAddressFinder
	s.LoginTemplate = template.Must(template.New("authPage").Parse(defLoginPage))

	cwd, err := crowd.New(user, password, crowdURL)
	if err != nil {
		return s, err
	}
	s.CrowdApp = &cwd
	s.CookieConfig, err = s.CrowdApp.GetCookieConfig()
	if err != nil {
		return s, err
	}

	return s, nil
}

// Handler provides HTTP middleware using http.Handler chaining
// that requires user authentication via Atlassian Crowd SSO.
func (s *SSO) Handler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.loginHandler(w, r) == false {
			return
		}
		h.ServeHTTP(w, r)
	})

}

func (s *SSO) loginHandler(w http.ResponseWriter, r *http.Request) bool {
	ck, err := r.Cookie(s.CookieConfig.Name)

	if err == http.ErrNoCookie {
		// no cookie so show login page if GET
		// if POST check if login and handle
		// if fail, show login page with message
		if r.Method == "GET" {
			s.LoginPage(w, r, s)
		} else if r.Method == "POST" {
			authOK := s.LoginPage(w, r, s)
			if authOK == true {
				// Redirect for fresh pass through auth etc on success
				http.Redirect(w, r, r.RequestURI, http.StatusTemporaryRedirect)
				return false
			} else {
				log.Printf("crowdauth: authentication failed\n")
			}
		} else {
			http.Error(w, authErr, http.StatusUnauthorized)
		}
		return false
	} else {
		// validate cookie or show login page
		host, err := s.ClientAddressFinder(r)
		if err != nil {
			log.Printf("crowdauth: could not get remote addr: %s\n", err)
			return false
		}

		_, err = s.CrowdApp.ValidateSession(ck.Value, host)
		if err != nil {
			log.Printf("crowdauth: could not validate cookie, deleting because: %s\n", err)
			s.EndSession(w, r)
			s.LoginPage(w, r, s)
			return false
		}

		// valid cookie so fallthrough
	}
	return true
}

func (s *SSO) Login(user string, pass string, addr string) (cs crowd.Session, err error) {
	cs, err = s.CrowdApp.NewSession(user, pass, addr)
	return cs, err
}

func (s *SSO) Logout(w http.ResponseWriter, r *http.Request, newURL string) {
	s.EndSession(w, r)
	http.Redirect(w, r, newURL, http.StatusTemporaryRedirect)
}

// StartSession sets a Crowd session cookie.
func (s *SSO) StartSession(w http.ResponseWriter, cs crowd.Session) {
	ck := http.Cookie{
		Name:    s.CookieConfig.Name,
		Domain:  s.CookieConfig.Domain,
		Secure:  s.CookieConfig.Secure,
		Value:   cs.Token,
		Expires: cs.Expires,
	}
	http.SetCookie(w, &ck)
}

// EndSession invalidates the current Crowd session and cookie
func (s *SSO) EndSession(w http.ResponseWriter, r *http.Request) {
	currentCookie, _ := r.Cookie(s.CookieConfig.Name)
	newCookie := &http.Cookie{
		Name:    s.CookieConfig.Name,
		Domain:  s.CookieConfig.Domain,
		Secure:  s.CookieConfig.Secure,
		MaxAge:  -1,
		Expires: time.Unix(1, 0),
		Value:   "LOGGED-OUT",
	}
	s.CrowdApp.InvalidateSession(currentCookie.Value)

	log.Printf("Got cookie to remove: %+v\n", currentCookie)
	log.Printf("Removal cookie: %+v\n", newCookie)
	http.SetCookie(w, newCookie)
}

// Get User information for the current session (by cookie)
func (s *SSO) GetUser(r *http.Request) (u crowd.User, err error) {
	currentCookie, err := r.Cookie(s.CookieConfig.Name)
	if err == http.ErrNoCookie {
		return u, errors.New("no session cookie")
	}

	userSession, err := s.CrowdApp.GetSession(currentCookie.Value)
	if err != nil {
		return u, errors.New("session not valid")
	}

	return userSession.User, nil
}

func loginPage(w http.ResponseWriter, r *http.Request, s *SSO) bool {
	if r.Method == "GET" { // show login page and bail
		showLoginPage(w, s)
		return false
	} else if r.Method == "POST" {
		user := r.FormValue("username")
		pass := r.FormValue("password")
		host, err := s.ClientAddressFinder(r)
		if err != nil {
			log.Printf("crowdauth: could not get remote addr: %s\n", err)
			showLoginPage(w, s)
			return false
		}

		sess, err := s.Login(user, pass, host)
		if err != nil {
			log.Printf("crowdauth: login/new session failed: %s\n", err)
			showLoginPage(w, s)
			return false
		}

		s.StartSession(w, sess)
	} else {
		return false
	}

	return true
}

func showLoginPage(w http.ResponseWriter, s *SSO) {
	err := s.LoginTemplate.ExecuteTemplate(w, "authPage", nil)
	if err != nil {
		log.Printf("crowdauth: could not exec template: %s\n", err)
	}
}
