# 0.25.0 / 2020-06-22

  * [#62] Add support for multiple domain requirements (#63)
  * Add cookie auth key environment variable (#59)

# 0.24.1 / 2020-04-08

  * Lint: Fix some minor linter errors
  * Fix: Config loading after CookieStore init (#58)

# 0.24.0 / 2020-01-13

  * [#50] Handle all 4xx errors as "user not found" (#52)

# 0.23.0 / 2019-12-28

  * Allow to configure anonymous access (#48)

# 0.22.0 / 2019-11-03

  * Switch to Go1.12+ vendoring
  * Fix: Broken HTML tag
  * Fix: Handle Unauthorized as no user found instead of generic error
  * Update vendored libraries

# 0.21.5 / 2019-06-29

  * [#41] Set default cookie values in all providers (#45)

# 0.21.4 / 2019-06-15

  * Prefer simple authenticator over LDAP (#42)

# 0.21.3 / 2019-05-14

  * Fix: Even with offline access no refresh token is present

# 0.21.2 / 2019-05-13

  * Fix: Google not returning refresh tokens

# 0.21.1 / 2019-04-26

  * Fix: Use cookie for redirects after oAuth flow

# 0.21.0 / 2019-04-23

  * [#35] Implement OpenID Connect auth provider
  * Fix: Only overwrite default if config is non-empty

# 0.20.1 / 2019-04-22

  * Fix: Do not list login methods without label

# 0.20.0 / 2019-04-22

  * Add special group for all authenticated users
  * Modernize login dialog

# 0.19.0 / 2019-04-22

  * Update dependencies
  * Move auth plugins to own modules
  * Move MFA plugins to own modules
  * Add default page in case neither redirect was specified
  * Implement oAuth2 provider: Google
  * Prepare moving auth plugins to own modules

# 0.18.0 / 2019-04-21

  * Add redirect on root URL to login page
  * Add default redirect URL for missing go-parameter

# 0.17.0 / 2019-04-21

  * Work around missing URL parameters (#39)

# 0.16.2 / 2019-04-16

  * Replace CDNJS as of permanent CORS failures

# 0.16.1 / 2019-03-17

  * Fix: Do not crash main program on incompatible plugins

# 0.16.0 / 2019-02-23

  * Enable CGO for plugin support
  * Add plugin support (#38)

# 0.15.1 / 2019-01-17

  * Fix: Host already had the port attached
  * Fix audit logging when not using MFA (#32)

# 0.15.0 / 2019-01-06

  * Add timestamp to audit log (#31)
  * Fix several linter errors

# 0.14.0 / 2018-12-29

  * [#25] Make TOTP provider fully configurable (#29)
  * Move documentation to project Wiki

# 0.13.0 / 2018-12-28

  * Add support for Duo MFA (#28)

# 0.12.0 / 2018-12-24

  * Implement MFA verification for logins (#10)

# 0.11.1 / 2018-11-18

  * [#19] Documentation improvements (#20)

# 0.11.0 / 2018-11-17

  * [#17] Implement audit logging

# 0.10.0 / 2018-09-24

  * Fix TLS dialing (#16)
  * Use multi-stage build to reduce image size

# 0.9.0 / 2018-09-20

  * Implement config reload on SIGHUP (#12)

# 0.8.1 / 2018-09-08

  * Fix: Memory leak due to http requests stored forever
  * Update repo-runner image

# 0.8.0 / 2018-07-26

  * Allow searching group members by username (#9)

# 0.7.1 / 2018-06-18

  * Fix: Ensure alias is set correctly when it is a DN

# 0.7.0 / 2018-06-18

  * Add configurable username to LDAP auth

# 0.6.0 / 2018-03-15

  * Add LDAP support (#3)

# 0.5.0 / 2018-02-04

  * Implement Crowd authentication (#2)

# 0.4.2 / 2018-02-04

  * Fix: Group assignments were not applied for Token auth

# 0.4.1 / 2018-02-04

  * Fix: Token auth always had a logged in user

# 0.4.0 / 2018-02-04

  * Allow grouping of tokens for simpler ACL

# 0.3.0 / 2018-01-28

  * Document auto-renewal
  * Auto-Renew cookies in simple and yubikey authenticators

# 0.2.0 / 2018-01-28

  * Add usage docs
  * Add basic auth to simple provider
  * Add dockerized version

# 0.1.0 / 2018-01-28

  * Initial version (#1)