package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/Luzifer/rconfig"
	"github.com/hashicorp/hcl"
	log "github.com/sirupsen/logrus"
)

type mainConfig struct {
	Cookie struct {
		Domain     string `hcl:"domain"`
		EncryptKey string `hcl:"encrypt_key"`
		Expire     int64  `hcl:"expire"`
		Prefix     string `hcl:"prefix"`
		Secure     bool   `hcl:"secure"`
	}
	Listen struct {
		Addr string `hcl:"addr"`
		Port int    `hcl:"port"`
	} `hcl:"listen"`
}

var (
	cfg = struct {
		ConfigFile     string `flag:"config,c" default:"config.hcl" env:"CONFIG" description:"Location of the configuration file"`
		TemplateDir    string `flag:"frontend-dir" default:"./frontend/" env:"FRONTEND_DIR" description:"Location of the directory containing the web assets"`
		VersionAndExit bool   `flag:"version" default:"false" description:"Prints current version and exits"`
	}{}

	mainCfg = mainConfig{}

	version = "dev"
)

func init() {
	if err := rconfig.Parse(&cfg); err != nil {
		log.Fatalf("Unable to parse commandline options: %s", err)
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

	http.ListenAndServe(fmt.Sprintf("%s:%d", mainCfg.Listen.Addr, mainCfg.Listen.Port), nil)
}
