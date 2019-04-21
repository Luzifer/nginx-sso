package main

import (
	"github.com/Luzifer/nginx-sso/plugins/auth/crowd"
	"github.com/Luzifer/nginx-sso/plugins/auth/google"
	"github.com/Luzifer/nginx-sso/plugins/auth/ldap"
	"github.com/Luzifer/nginx-sso/plugins/auth/simple"
	"github.com/Luzifer/nginx-sso/plugins/auth/token"
	auth_yubikey "github.com/Luzifer/nginx-sso/plugins/auth/yubikey"
	"github.com/Luzifer/nginx-sso/plugins/mfa/duo"
	"github.com/Luzifer/nginx-sso/plugins/mfa/totp"
	mfa_yubikey "github.com/Luzifer/nginx-sso/plugins/mfa/yubikey"
)

func registerModules() {
	registerAuthenticator(crowd.New())
	registerAuthenticator(ldap.New(cookieStore))
	registerAuthenticator(google.New(cookieStore))
	registerAuthenticator(simple.New(cookieStore))
	registerAuthenticator(token.New())
	registerAuthenticator(auth_yubikey.New(cookieStore))

	registerMFAProvider(duo.New())
	registerMFAProvider(totp.New())
	registerMFAProvider(mfa_yubikey.New())
}
