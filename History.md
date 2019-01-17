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