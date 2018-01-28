package rconfig

import (
	"fmt"
	"os"
)

func ExampleParse() {
	// We're building an example configuration with a sub-struct to be filled
	// by the Parse command.
	config := struct {
		Username string `default:"unknown" flag:"user,u" description:"Your name"`
		Details  struct {
			Age int `default:"25" flag:"age" description:"Your age"`
		}
	}{}

	// To have more relieable results we're setting os.Args to a known value.
	// In real-life use cases you wouldn't do this but parse the original
	// commandline arguments.
	os.Args = []string{
		"example",
		"--user=Luzifer",
	}

	Parse(&config)

	fmt.Printf("Hello %s, happy birthday for your %dth birthday.",
		config.Username,
		config.Details.Age)

	// You can also show an usage message for your user
	Usage()

	// Output:
	// Hello Luzifer, happy birthday for your 25th birthday.
}
