package plugins

type RegisterAuthenticatorFunc func(Authenticator)
type RegisterMFAProviderFunc func(MFAProvider)
