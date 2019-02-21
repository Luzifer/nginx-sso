package main

import (
	"os"
	"path/filepath"
	"plugin"
	"strings"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/Luzifer/nginx-sso/plugins"
)

func loadPlugins(pluginDir string) error {
	logger := log.WithField("plugin_dir", pluginDir)

	d, err := os.Stat(pluginDir)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Warn("Plugin directory not found, skipping")
			return nil
		}
		return errors.Wrap(err, "Could not stat plugin dir")
	}

	if !d.IsDir() {
		return errors.New("Plugin directory is not a directory")
	}

	return errors.Wrap(filepath.Walk(pluginDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !strings.HasSuffix(path, ".so") {
			// Ignore that file, is not a plugin
			return nil
		}

		p, err := plugin.Open(path)
		if err != nil {
			return errors.Wrapf(err, "Unable to load plugin %q", path)
		}

		f, err := p.Lookup("Register")
		if err != nil {
			return errors.Wrapf(err, "Unable to find register function in %q", path)
		}

		f.(func(plugins.RegisterAuthenticatorFunc, plugins.RegisterMFAProviderFunc))(
			registerAuthenticator,
			registerMFAProvider,
		)

		return nil
	}), "Unable to load plugins")
}
