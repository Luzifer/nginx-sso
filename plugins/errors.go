package plugins

import "errors"

var (
	ErrProviderUnconfigured = errors.New("No valid configuration found for this provider")
	ErrNoValidUserFound     = errors.New("No valid users found")
)
