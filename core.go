package main

import "github.com/Luzifer/nginx-sso/plugins/auth/google"

func registerModules() {
	registerAuthenticator(google.New(cookieStore))
}
