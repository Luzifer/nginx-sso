package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"

	"github.com/flosch/pongo2"
	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"

	"github.com/Luzifer/nginx-sso/plugins"
	"github.com/Luzifer/rconfig"
)

type mainConfig struct {
	ACL      acl                  `yaml:"acl"`
	AuditLog auditLogger          `yaml:"audit_log"`
	Cookie   plugins.CookieConfig `yaml:"cookie"`
	Listen   struct {
		Addr string `yaml:"addr"`
		Port int    `yaml:"port"`
	} `yaml:"listen"`
	Login struct {
		Title           string            `yaml:"title"`
		DefaultMethod   string            `yaml:"default_method"`
		DefaultRedirect string            `yaml:"default_redirect"`
		HideMFAField    bool              `yaml:"hide_mfa_field"`
		Names           map[string]string `yaml:"names"`
	} `yaml:"login"`
	Plugins struct {
		Directory string `yaml:"directory"`
	} `yaml:"plugins"`
}

var (
	cfg = struct {
		ConfigFile     string `flag:"config,c" default:"config.yaml" env:"CONFIG" description:"Location of the configuration file"`
		LogLevel       string `flag:"log-level" default:"info" description:"Level of logs to display (debug, info, warn, error)"`
		TemplateDir    string `flag:"frontend-dir" default:"./frontend/" env:"FRONTEND_DIR" description:"Location of the directory containing the web assets"`
		VersionAndExit bool   `flag:"version" default:"false" description:"Prints current version and exits"`
	}{}

	mainCfg     = mainConfig{}
	cookieStore *sessions.CookieStore

	version = "dev"
)

func init() {
	if err := rconfig.Parse(&cfg); err != nil {
		log.WithError(err).Fatal("Unable to parse commandline options")
	}

	if l, err := log.ParseLevel(cfg.LogLevel); err != nil {
		log.WithError(err).Fatal("Unable to parse log level")
	} else {
		log.SetLevel(l)
	}

	if cfg.VersionAndExit {
		fmt.Printf("nginx-sso %s\n", version)
		os.Exit(0)
	}

	// Set sane defaults for main configuration
	mainCfg.Cookie.Prefix = "nginx-sso"
	mainCfg.Cookie.Expire = 3600
	mainCfg.Listen.Addr = "127.0.0.1"
	mainCfg.Listen.Port = 8082
	mainCfg.Login.DefaultRedirect = "debug"
	mainCfg.AuditLog.TrustedIPHeaders = []string{"X-Forwarded-For", "RemoteAddr", "X-Real-IP"}
	mainCfg.AuditLog.Headers = []string{"x-origin-uri"}
}

func loadConfiguration() error {
	yamlSource, err := ioutil.ReadFile(cfg.ConfigFile)
	if err != nil {
		return fmt.Errorf("Unable to read configuration file: %s", err)
	}

	if err = yaml.Unmarshal(yamlSource, &mainCfg); err != nil {
		return fmt.Errorf("Unable to load configuration file: %s", err)
	}

	if mainCfg.Plugins.Directory != "" {
		if err = loadPlugins(mainCfg.Plugins.Directory); err != nil {
			return errors.Wrap(err, "Unable to load plugins")
		}
	}

	if err = initializeAuthenticators(yamlSource); err != nil {
		return fmt.Errorf("Unable to configure authentication: %s", err)
	}

	if err = initializeMFAProviders(yamlSource); err != nil {
		log.WithError(err).Fatal("Unable to configure MFA providers")
	}

	return nil
}

func main() {
	cookieStore = sessions.NewCookieStore([]byte(mainCfg.Cookie.AuthKey))
	registerModules()

	if err := loadConfiguration(); err != nil {
		log.WithError(err).Fatal("Unable to load configuration")
	}

	http.HandleFunc("/", handleRootRequest)
	http.HandleFunc("/auth", handleAuthRequest)
	http.HandleFunc("/debug", handleLoginDebug)
	http.HandleFunc("/login", handleLoginRequest)
	http.HandleFunc("/logout", handleLogoutRequest)

	go http.ListenAndServe(
		fmt.Sprintf("%s:%d", mainCfg.Listen.Addr, mainCfg.Listen.Port),
		context.ClearHandler(http.DefaultServeMux),
	)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGHUP)

	for sig := range sigChan {
		switch sig {
		case syscall.SIGHUP:
			if err := loadConfiguration(); err != nil {
				log.WithError(err).Error("Unable to reload configuration")
			}

		default:
			log.Fatalf("Received unexpected signal: %v", sig)
		}
	}
}

func handleRootRequest(res http.ResponseWriter, r *http.Request) {
	// In case of a request to `/` redirect to login utilizing the default redirect
	http.Redirect(res, r, "login", http.StatusFound)
}

func handleAuthRequest(res http.ResponseWriter, r *http.Request) {
	user, groups, err := detectUser(res, r)

	switch err {
	case plugins.ErrNoValidUserFound:
		mainCfg.AuditLog.Log(auditEventValidate, r, map[string]string{"result": "no valid user found"}) // #nosec G104 - This is only logging
		http.Error(res, "No valid user found", http.StatusUnauthorized)

	case nil:
		if !mainCfg.ACL.HasAccess(user, groups, r) {
			mainCfg.AuditLog.Log(auditEventAccessDenied, r, map[string]string{"username": user}) // #nosec G104 - This is only logging
			http.Error(res, "Access denied for this resource", http.StatusForbidden)
			return
		}

		mainCfg.AuditLog.Log(auditEventValidate, r, map[string]string{"result": "valid user found", "username": user}) // #nosec G104 - This is only logging

		res.Header().Set("X-Username", user)
		res.WriteHeader(http.StatusOK)

	default:
		log.WithError(err).Error("Error while handling auth request")
		http.Error(res, "Something went wrong", http.StatusInternalServerError)
	}
}

func handleLoginRequest(res http.ResponseWriter, r *http.Request) {
	redirURL, err := getRedirectURL(r, mainCfg.Login.DefaultRedirect)
	if err != nil {
		http.Error(res, "Invalid redirect URL specified", http.StatusBadRequest)
	}

	if _, _, err := detectUser(res, r); err == nil {
		// There is already a valid user
		http.Redirect(res, r, redirURL, http.StatusFound)
		return
	}

	auditFields := map[string]string{
		"go": redirURL,
	}

	if r.Method == "POST" || r.URL.Query().Get("code") != "" {
		// Simple authentication
		user, mfaCfgs, err := loginUser(res, r)
		switch err {
		case plugins.ErrNoValidUserFound:
			auditFields["reason"] = "invalid credentials"
			mainCfg.AuditLog.Log(auditEventLoginFailure, r, auditFields) // #nosec G104 - This is only logging
			http.Redirect(res, r, "/login?go="+url.QueryEscape(redirURL), http.StatusFound)
			return
		case nil:
			// Don't handle for now, MFA validation comes first
		default:
			auditFields["reason"] = "error"
			auditFields["error"] = err.Error()
			mainCfg.AuditLog.Log(auditEventLoginFailure, r, auditFields) // #nosec G104 - This is only logging
			log.WithError(err).Error("Login failed with unexpected error")
			http.Redirect(res, r, "/login?go="+url.QueryEscape(redirURL), http.StatusFound)
			return
		}

		// MFA validation against configs from login
		err = validateMFA(res, r, user, mfaCfgs)
		switch err {
		case plugins.ErrNoValidUserFound:
			auditFields["reason"] = "invalid credentials"
			mainCfg.AuditLog.Log(auditEventLoginFailure, r, auditFields) // #nosec G104 - This is only logging
			res.Header().Del("Set-Cookie")                               // Remove login cookie
			http.Redirect(res, r, "/login?go="+url.QueryEscape(redirURL), http.StatusFound)
			return

		case nil:
			mainCfg.AuditLog.Log(auditEventLoginSuccess, r, auditFields) // #nosec G104 - This is only logging
			http.Redirect(res, r, redirURL, http.StatusFound)
			return

		default:
			auditFields["reason"] = "error"
			auditFields["error"] = err.Error()
			mainCfg.AuditLog.Log(auditEventLoginFailure, r, auditFields) // #nosec G104 - This is only logging
			log.WithError(err).Error("Login failed with unexpected error")
			res.Header().Del("Set-Cookie") // Remove login cookie
			http.Redirect(res, r, "/login?go="+url.QueryEscape(redirURL), http.StatusFound)
			return
		}
	}

	// Store redirect URL in session (required for oAuth2 flows)
	sess, _ := cookieStore.Get(r, strings.Join([]string{mainCfg.Cookie.Prefix, "main"}, "-")) // #nosec G104 - On error empty session is returned
	sess.Options = mainCfg.Cookie.GetSessionOpts()
	sess.Values["go"] = redirURL

	if err := sess.Save(r, res); err != nil {
		log.WithError(err).Error("Unable to save session")
		http.Error(res, "Something went wrong", http.StatusInternalServerError)
	}

	// Render login page
	tpl := pongo2.Must(pongo2.FromFile(path.Join(cfg.TemplateDir, "index.html")))
	if err := tpl.ExecuteWriter(pongo2.Context{
		"active_methods": getFrontendAuthenticators(),
		"go":             redirURL,
		"login":          mainCfg.Login,
	}, res); err != nil {
		log.WithError(err).Error("Unable to render template")
		http.Error(res, "Something went wrong", http.StatusInternalServerError)
	}
}

func handleLogoutRequest(res http.ResponseWriter, r *http.Request) {
	redirURL, err := getRedirectURL(r, mainCfg.Login.DefaultRedirect)
	if err != nil {
		http.Error(res, "Invalid redirect URL specified", http.StatusBadRequest)
	}

	mainCfg.AuditLog.Log(auditEventLogout, r, nil) // #nosec G104 - This is only logging
	if err := logoutUser(res, r); err != nil {
		log.WithError(err).Error("Failed to logout user")
		http.Error(res, "Something went wrong", http.StatusInternalServerError)
		return
	}

	http.Redirect(res, r, redirURL, http.StatusFound)
}

func handleLoginDebug(w http.ResponseWriter, r *http.Request) {
	user, groups, err := detectUser(w, r)
	switch err {
	case nil:
		// All fine

	case plugins.ErrNoValidUserFound:
		http.Redirect(w, r, "login", http.StatusFound)
		return

	default:
		log.WithError(err).Error("Failed to get user for login debug")
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Successfully logged in:")
	fmt.Fprintf(w, "- Username: %s\n", user)
	fmt.Fprintf(w, "- Groups: %s\n", strings.Join(groups, ","))
}
