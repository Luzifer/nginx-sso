package which

import (
	"errors"
	"os"
	"path"
	"strings"
)

// Common named errors to match in programs using this library
var (
	ErrBinaryNotFound    = errors.New("Requested binary was not found")
	ErrNoSearchSpecified = errors.New("You need to specify a binary to search")
)

// FindInPath searches the specified binary in directories listed in $PATH and returns first match
func FindInPath(binary string) (string, error) {
	pathEnv := os.Getenv("PATH")
	if len(pathEnv) == 0 {
		return "", errors.New("Found empty $PATH, not able to search $PATH")
	}

	for _, part := range strings.Split(pathEnv, ":") {
		if len(part) == 0 {
			continue
		}

		if found, err := FindInDirectory(binary, part); err != nil {
			return "", err
		} else if found {
			return path.Join(part, binary), nil
		}
	}

	return "", ErrBinaryNotFound
}

// FindInDirectory checks whether the specified file is present in the directory
func FindInDirectory(binary, directory string) (bool, error) {
	if len(binary) == 0 {
		return false, ErrNoSearchSpecified
	}

	_, err := os.Stat(path.Join(directory, binary))

	switch {
	case err == nil:
		return true, nil
	case os.IsNotExist(err):
		return false, nil
	default:
		return false, err
	}
}
