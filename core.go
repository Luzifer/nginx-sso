package main

import (
	"github.com/Luzifer/nginx-sso/plugins/auth/google"
	"github.com/Luzifer/nginx-sso/plugins/mfa/duo"
	"github.com/Luzifer/nginx-sso/plugins/mfa/totp"
	mfa_yubikey "github.com/Luzifer/nginx-sso/plugins/mfa/yubikey"
)

func registerModules() {
	registerAuthenticator(google.New(cookieStore))

	registerMFAProvider(duo.New())
	registerMFAProvider(totp.New())
	registerMFAProvider(mfa_yubikey.New())
}
