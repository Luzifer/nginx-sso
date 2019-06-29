package plugins

import "github.com/gorilla/sessions"

type CookieConfig struct {
	Domain  string `yaml:"domain"`
	AuthKey string `yaml:"authentication_key"`
	Expire  int    `yaml:"expire"`
	Prefix  string `yaml:"prefix"`
	Secure  bool   `yaml:"secure"`
}

func (c CookieConfig) GetSessionOpts() *sessions.Options {
	return &sessions.Options{
		Path:     "/",
		Domain:   c.Domain,
		MaxAge:   c.Expire,
		Secure:   c.Secure,
		HttpOnly: true,
	}
}

func DefaultCookieConfig() CookieConfig {
	return CookieConfig{
		Prefix: "nginx-sso",
		Expire: 3600,
	}
}
