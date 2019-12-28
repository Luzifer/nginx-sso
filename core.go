package main

import (
	"github.com/Luzifer/nginx-sso/plugins/auth/crowd"
	"github.com/Luzifer/nginx-sso/plugins/auth/google"
	"github.com/Luzifer/nginx-sso/plugins/auth/ldap"
	"github.com/Luzifer/nginx-sso/plugins/auth/oidc"
	"github.com/Luzifer/nginx-sso/plugins/auth/simple"
	"github.com/Luzifer/nginx-sso/plugins/auth/token"
	auth_yubikey "github.com/Luzifer/nginx-sso/plugins/auth/yubikey"
	"github.com/Luzifer/nginx-sso/plugins/mfa/duo"
	"github.com/Luzifer/nginx-sso/plugins/mfa/totp"
	mfa_yubikey "github.com/Luzifer/nginx-sso/plugins/mfa/yubikey"
)

func registerModules() {
	// Start with very simple, local auth providers as they are cheap
	// in their execution and therefore if they are used nginx-sso
	// can process far more requests than through the other providers
	registerAuthenticator(simple.New(cookieStore))
	registerAuthenticator(token.New())

	// Afterwards utilize the more expensive remove providers
	registerAuthenticator(crowd.New())
	registerAuthenticator(ldap.New(cookieStore))
	registerAuthenticator(google.New(cookieStore))
	registerAuthenticator(oidc.New(cookieStore))
	registerAuthenticator(auth_yubikey.New(cookieStore))

	registerMFAProvider(duo.New())
	registerMFAProvider(totp.New())
	registerMFAProvider(mfa_yubikey.New())
}
