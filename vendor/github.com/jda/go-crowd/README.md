go-crowd
========
Go library for interacting with [Atlassian Crowd](https://www.atlassian.com/software/crowd/)

* [![GoDoc](https://godoc.org/github.com/jda/go-crowd?status.png)](http://godoc.org/github.com/jda/go-crowd)
* Crowd [API Documentation](https://developer.atlassian.com/display/CROWDDEV/Remote+API+Reference)

## Client example
```go
client, err := crowd.New("crowd_app_user", 
                        "crowd_app_password", 
                        "crowd service URL")

user, err := client.Authenticate("testuser", "testpass")
if err != nil {
    /*
    failure or reject from crowd. check if err = reason from 
    https://developer.atlassian.com/display/CROWDDEV/Using+the+Crowd+REST+APIs#UsingtheCrowdRESTAPIs-HTTPResponseCodesandErrorResponses
    */
}

// if auth successful, user contains user information
```