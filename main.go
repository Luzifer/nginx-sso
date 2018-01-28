package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"

	"github.com/Luzifer/rconfig"
	"github.com/flosch/pongo2"
	"github.com/gorilla/sessions"
	"github.com/hashicorp/hcl"
	log "github.com/sirupsen/logrus"
)

type mainConfig struct {
	Cookie struct {
		Domain  string `hcl:"domain"`
		AuthKey string `hcl:"authentication_key"`
		Expire  int    `hcl:"expire"`
		Prefix  string `hcl:"prefix"`
		Secure  bool   `hcl:"secure"`
	}
	Listen struct {
		Addr string `hcl:"addr"`
		Port int    `hcl:"port"`
	} `hcl:"listen"`
	Login struct {
		Title         string            `hcl:"title"`
		DefaultMethod string            `hcl:"default_method"`
		Names         map[string]string `hcl:"names"`
	} `hcl:"login"`
}

var (
	cfg = struct {
		ConfigFile     string `flag:"config,c" default:"config.hcl" env:"CONFIG" description:"Location of the configuration file"`
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
}

func main() {
	hclSource, err := ioutil.ReadFile(cfg.ConfigFile)
	if err != nil {
		log.WithError(err).Fatal("Unable to read configuration file")
	}

	if err := hcl.Unmarshal(hclSource, &mainCfg); err != nil {
		log.WithError(err).Fatal("Unable to load configuration file")
	}

	if err := initializeAuthenticators(hclSource); err != nil {
		log.WithError(err).Fatal("Unable to configure authentication")
	}

	cookieStore = sessions.NewCookieStore([]byte(mainCfg.Cookie.AuthKey))

	http.HandleFunc("/auth", handleAuthRequest)
	http.HandleFunc("/login", handleLoginRequest)
	http.HandleFunc("/logout", handleLogoutRequest)

	http.ListenAndServe(fmt.Sprintf("%s:%d", mainCfg.Listen.Addr, mainCfg.Listen.Port), nil)
}

func handleAuthRequest(res http.ResponseWriter, r *http.Request) {
	user, groups, err := detectUser(r)

	switch err {
	case noValidUserFoundError:
		http.Error(res, "No valid user found", http.StatusUnauthorized)

	case nil:
		// FIXME (kahlers): Do ACL check here and decide to answer 403
		_ = groups

		res.Header().Set("X-Username", user)
		res.WriteHeader(http.StatusOK)

	default:
		log.WithError(err).Error("Error while handling auth request")
		http.Error(res, "Something went wrong", http.StatusInternalServerError)
	}
}

func handleLoginRequest(res http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		err := loginUser(res, r)
		switch err {
		case noValidUserFoundError:
			http.Redirect(res, r, "/login?go="+url.QueryEscape(r.FormValue("go")), http.StatusFound)
			return
		case nil:
			http.Redirect(res, r, r.FormValue("go"), http.StatusFound)
			return
		default:
			log.WithError(err).Error("Login failed with unexpected error")
			http.Redirect(res, r, "/login?go="+url.QueryEscape(r.FormValue("go")), http.StatusFound)
			return
		}
	}

	tpl := pongo2.Must(pongo2.FromFile(path.Join(cfg.TemplateDir, "index.html")))
	if err := tpl.ExecuteWriter(pongo2.Context{
		"active_methods": getFrontendAuthenticators(),
		"login":          mainCfg.Login,
	}, res); err != nil {
		log.WithError(err).Error("Unable to render template")
		http.Error(res, "Something went wrong", http.StatusInternalServerError)
	}
}

func handleLogoutRequest(res http.ResponseWriter, r *http.Request) {}
