# Luzifer / nginx-sso

This program is intended to be used within the [`ngx_http_auth_request_module`](https://nginx.org/en/docs/http/ngx_http_auth_request_module.html) of nginx to provide a single-sign-on for a domain using one central authentication directory.

## Configuration

The configuration is mainly done using a YAML configuration file. Some options are configurable through command line flags and can be looked up using `--help` flag.

For an example configuration see the [`config.yaml`](config.yaml) file in this repository. Within the next sections the options are explained in more detail:

### Main configuration: Login form

The login form can be customized with its wording and the default login method.

```yaml
login:
  title: "luzifer.io - Login"
  default_method: "simple"
  names:
    simple: "Username / Password"
    yubikey: "Yubikey"
```

Most options should explain themselves, the `names` dictionary maps IDs of the authentication methods (shown in the title of their config section below) to human readable strings. You can set any string you need and your user recognizes.

### Main configuration: Cookie Settings

Most of the cookie settings are pre-set to sane defaults but you definitly need to configure some.

```yaml
cookie:
  domain: ".example.com"
  authentication_key: "Ff1uWJcLouKu9kwxgbnKcU3ps47gps72sxEz79TGHFCpJNCPtiZAFDisM4MWbstH"
  expire: 3600        # Optional, default: 3600
  prefix: "nginx-sso" # Optional, default: nginx-sso
  secure: true        # Optional, default: false
```

Adjust the `domain` to your service. So if all of your services live under `*.luzifer.io` you want to set the domain to `.luzifer.io`. The `authentication_key` needs to be set to some unique string not known to others. It is used to validate nobody messed with your session cookies. If this is leaked (or you just used the default) attackers can just set any username inside the corresponding cookie and are able to access your services!

If you are accessing your services through HTTPs you want to enable `secure` cookies. Also you should think about customizing the cookie `prefix` and the `expire` time of the cookie.

### Main configuration: HTTP Listener

This section configures where you can reach the program using HTTP and where you will point your nginx to. The example below shows the defaults and you don't need to change them.

```yaml
listen:
  addr: "127.0.0.1"
  port: 8082
```

### Main configuration: ACL

The rules of the ACL are the most complex part of the configuration and you should take your time to make this bullet-proof. If you mess up you're probably are getting complaints from your users because the default policy applied is to `deny` all access. So in the end you are configuring a white-list here.

```yaml
acl:
  rule_sets:
  - rules:
    - field: "host"
      equals: "test.example.com"
    - field: "x-origin-uri"
      regexp: "^/api"
    allow: ["luzifer", "@admins"]
```

Each `rule_sets` entry consists of three parts: `rules`, `allow` and `deny` directives. You can supply as many rules as you need, they are connected using AND logic per rule-set.

Each `rules` entry has two mandantory and three optional fields of which at least one *must* be set:
- `field` - required - Selector of the header your nginx is sending to the `/auth` endpoint (e.g. `Host`, `X-Origin-URI`, ...)
- `invert` - required - Boolean used to invert the matching: What was true will be false. Useful for "does not match this regexp" rules (default: `false`)
- `present` - optional - Boolean stating a certain header must exist or must not exist
- `regexp` - optional - String containing a regexp which must match the contents of the header selected by `field`
- `equals` - optional - String which must fully match the contents of the header selected by `field`

The `allow` and `deny` directives are arrays of users and groups. Groups are prefixed using an `@` sign. There is a simple logic: Users before groups, denies before allows. So if you allow the group `@test` containing the user `mike` but deny the user `mike`, mike will not be able to access the matching sites.

### Provider configuration: Simple Auth (`simple`)

The simple auth provider consists of a static mapping between users and passwords and groups and users. This can be seen as the replacement of htpasswd files.

```yaml
providers:
  simple:
    # Unique username mapped to bcrypt hashed password
    users:
      luzifer: "$2a$10$FSGAF8qDWX52aBID8.WpxOyCvfSQ3JIUVFiwyd1jolb4jM3BzJmNu"
      mike: "$2a$10$/0nrpYkdVhAifCLCI1DTz.4CkbCkc8CsvYhfvBRIhTTQDfBrkJ8Re"

    # Groupname to users mapping
    groups:
      admins: ["luzifer"]
      users: ["mike"]
```

You can see how to configure the provider the example above: No surprises, just ensure you are using bcrypt hashes for the passwords, no other hash functions are supported.

### Provider configuration: Token Auth (`token`)

The token auth provider is intended to give machines access to endpoints. Users will not be able to "login" using tokens when they see the login form.

```yaml
providers:
  token:
    # Mapping of unique token names to the token
    tokens:
      tokenname: "MYTOKEN"
      mycli: "kQHjQLuQdkSPwdJ1mueniLMPSjCc6GVt"
```

This provider does not support grouping: Each token needs to be white-listed explicitly. When accessing the sites using a token this header is expected:

`Authorization: Token MYTOKEN`

### Provider configuration: Yubikey One-Factor-Auth (`yubikey`)

The Yubikey auth provider is a one-factor-authentication mechanism. Not to be confused by U2F or HOTP two-factor methods. Your users only need to press the button to fully login. (Be sure you know what you're doing here!)

```yaml
providers:
  yubikey:
    # Get your client / secret from https://upgrade.yubico.com/getapikey/
    client_id: "12345"
    secret_key: "foobar"

    # First 12 characters of the OTP string mapped to the username
    devices:
      ccccccfcvuul: "luzifer"

    # Groupname to users mapping
    groups:
      admins: ["luzifer"]
```

You need to configure the `client_id` and the `secret_key` for the Yubico online validation service and the Yubikeys need to comply the specifications of that API (do not put random values into the device ID). Afterwards just take the first 12 characters of the keys OTP and map it to an user.
