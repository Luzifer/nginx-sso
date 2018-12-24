package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path"
	"syscall"

	"github.com/flosch/pongo2"
	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"

	"github.com/Luzifer/rconfig"
)

type mainConfig struct {
	ACL      acl         `yaml:"acl"`
	AuditLog auditLogger `yaml:"audit_log"`
	Cookie   struct {
		Domain  string `yaml:"domain"`
		AuthKey string `yaml:"authentication_key"`
		Expire  int    `yaml:"expire"`
		Prefix  string `yaml:"prefix"`
		Secure  bool   `yaml:"secure"`
	}
	Listen struct {
		Addr string `yaml:"addr"`
		Port int    `yaml:"port"`
	} `yaml:"listen"`
	Login struct {
		Title         string            `yaml:"title"`
		DefaultMethod string            `yaml:"default_method"`
		HideMFAField  bool              `yaml:"hide_mfa_field"`
		Names         map[string]string `yaml:"names"`
	} `yaml:"login"`
}

func (m *mainConfig) GetSessionOpts() *sessions.Options {
	return &sessions.Options{
		Path:     "/",
		Domain:   m.Cookie.Domain,
		MaxAge:   m.Cookie.Expire,
		Secure:   m.Cookie.Secure,
		HttpOnly: true,
	}
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
	mainCfg.AuditLog.TrustedIPHeaders = []string{"X-Forwarded-For", "RemoteAddr", "X-Real-IP"}
	mainCfg.AuditLog.Headers = []string{"x-origin-uri"}
}

func loadConfiguration() error {
	yamlSource, err := ioutil.ReadFile(cfg.ConfigFile)
	if err != nil {
		return fmt.Errorf("Unable to read configuration file: %s", err)
	}

	if err := yaml.Unmarshal(yamlSource, &mainCfg); err != nil {
		return fmt.Errorf("Unable to load configuration file: %s", err)
	}

	if err := initializeAuthenticators(yamlSource); err != nil {
		return fmt.Errorf("Unable to configure authentication: %s", err)
	}

	if err = initializeMFAProviders(yamlSource); err != nil {
		log.WithError(err).Fatal("Unable to configure MFA providers")
	}

	return nil
}

func main() {
	if err := loadConfiguration(); err != nil {
		log.WithError(err).Fatal("Unable to load configuration")
	}

	cookieStore = sessions.NewCookieStore([]byte(mainCfg.Cookie.AuthKey))

	http.HandleFunc("/auth", handleAuthRequest)
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

func handleAuthRequest(res http.ResponseWriter, r *http.Request) {
	user, groups, err := detectUser(res, r)

	switch err {
	case errNoValidUserFound:
		mainCfg.AuditLog.Log(auditEventValidate, r, map[string]string{"result": "no valid user found"})
		http.Error(res, "No valid user found", http.StatusUnauthorized)

	case nil:
		if !mainCfg.ACL.HasAccess(user, groups, r) {
			mainCfg.AuditLog.Log(auditEventAccessDenied, r, map[string]string{"username": user})
			http.Error(res, "Access denied for this resource", http.StatusForbidden)
			return
		}

		mainCfg.AuditLog.Log(auditEventValidate, r, map[string]string{"result": "valid user found", "username": user})

		res.Header().Set("X-Username", user)
		res.WriteHeader(http.StatusOK)

	default:
		log.WithError(err).Error("Error while handling auth request")
		http.Error(res, "Something went wrong", http.StatusInternalServerError)
	}
}

func handleLoginRequest(res http.ResponseWriter, r *http.Request) {
	if _, _, err := detectUser(res, r); err == nil {
		// There is already a valid user
		http.Redirect(res, r, r.URL.Query().Get("go"), http.StatusFound)
		return
	}

	auditFields := map[string]string{
		"go": r.FormValue("go"),
	}

	if r.Method == "POST" {
		// Simple authentication
		user, mfaCfgs, err := loginUser(res, r)
		switch err {
		case errNoValidUserFound:
			http.Redirect(res, r, "/login?go="+url.QueryEscape(r.FormValue("go")), http.StatusFound)
			return
		case nil:
			// Don't handle for now, MFA validation comes first
		default:
			log.WithError(err).Error("Login failed with unexpected error")
			http.Redirect(res, r, "/login?go="+url.QueryEscape(r.FormValue("go")), http.StatusFound)
			return
		}

		// MFA validation against configs from login
		err = validateMFA(res, r, user, mfaCfgs)
		switch err {
		case errNoValidUserFound:
			auditFields["reason"] = "invalid credentials"
			mainCfg.AuditLog.Log(auditEventLoginFailure, r, auditFields)
			res.Header().Del("Set-Cookie") // Remove login cookie
			http.Redirect(res, r, "/login?go="+url.QueryEscape(r.FormValue("go")), http.StatusFound)
			return

		case nil:
			mainCfg.AuditLog.Log(auditEventLoginSuccess, r, auditFields)
			http.Redirect(res, r, r.FormValue("go"), http.StatusFound)
			return

		default:
			auditFields["reason"] = "error"
			auditFields["error"] = err.Error()
			mainCfg.AuditLog.Log(auditEventLoginFailure, r, auditFields)
			log.WithError(err).Error("Login failed with unexpected error")
			res.Header().Del("Set-Cookie") // Remove login cookie
			http.Redirect(res, r, "/login?go="+url.QueryEscape(r.FormValue("go")), http.StatusFound)
			return
		}
	}

	tpl := pongo2.Must(pongo2.FromFile(path.Join(cfg.TemplateDir, "index.html")))
	if err := tpl.ExecuteWriter(pongo2.Context{
		"active_methods": getFrontendAuthenticators(),
		"go":             r.URL.Query().Get("go"),
		"login":          mainCfg.Login,
	}, res); err != nil {
		log.WithError(err).Error("Unable to render template")
		http.Error(res, "Something went wrong", http.StatusInternalServerError)
	}
}

func handleLogoutRequest(res http.ResponseWriter, r *http.Request) {
	mainCfg.AuditLog.Log(auditEventLogout, r, nil)
	if err := logoutUser(res, r); err != nil {
		log.WithError(err).Error("Failed to logout user")
		http.Error(res, "Something went wrong", http.StatusInternalServerError)
		return
	}

	http.Redirect(res, r, r.URL.Query().Get("go"), http.StatusFound)
}
