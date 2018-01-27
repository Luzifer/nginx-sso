cookie {
  domain      = ".example.com"
  encrypt_key = "Ff1uWJcLouKu9kwxgbnKcU3ps47gps72sxEz79TGHFCpJNCPtiZAFDisM4MWbstH"
  expire      = 3600                                                               // Optional, default: 3600
  prefix      = "nginx-sso"                                                        // Optional, default: nginx-sso
  secure      = true                                                               // Optional, default: false
}

// Optional, default: 127.0.0.1:8082
listen {
  addr = "127.0.0.1"
  port = 8082
}

providers {
  // Authentication against an Atlassian Crowd directory server
  // Supports: Users, Groups
  crowd {
    url  = "https://crowd.example.com/crowd/"
    user = ""
    pass = ""
  }

  // Authentication against embedded user database
  // Supports: Users, Groups
  simple {
    // Unique username mapped to bcrypt hashed password
    users {
      luzifer = "$2a$10$V0X4fp9B9TE2woDzhj3pVunxno1M0RtpHaxeDdo0AKrUxUN8s6IIi"
    }

    // Groupname to users mapping
    groups {
      admins = ["luzifer"]
    }
  }

  // Authentication against embedded token directory
  // Supports: Users
  token {
    // Mapping of unique token names to the token
    tokens {
      tokenname = "MYTOKEN"
    }
  }

  // Authentication against Yubikey cloud validation servers
  // Supports: Users, Groups
  yubikey {
    api_key            = "foobar"
    validation_servers = ["myserver.example.com"] // Optional, defaults to Yubico cloud servers

    // First 12 characters of the OTP string mapped to the username
    devices {
      ccccccfcvuul = "luzifer"
    }

    // Groupname to users mapping
    groups {
      admins = ["luzifer"]
    }
  }
}
