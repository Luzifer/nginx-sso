package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/Luzifer/go_helpers/v2/str"
)

type auditEvent string

const (
	auditEventAccessDenied            = "access_denied"
	auditEventLoginFailure            = "login_failure"
	auditEventLoginSuccess auditEvent = "login_success"
	auditEventLogout                  = "logout"
	auditEventValidate                = "validate"
)

type auditLogger struct {
	Targets          []string `yaml:"targets"`
	Events           []string `yaml:"events"`
	Headers          []string `yaml:"headers"`
	TrustedIPHeaders []string `yaml:"trusted_ip_headers"`

	lock sync.Mutex
}

func (a *auditLogger) Log(event auditEvent, r *http.Request, extraFields map[string]string) error {
	if len(a.Targets) == 0 {
		return nil
	}

	if !str.StringInSlice(string(event), a.Events) {
		return nil
	}

	// Ensure order of logs, prevent file operation collisions
	a.lock.Lock()
	defer a.lock.Unlock()

	// Compile log event
	evt := map[string]interface{}{}
	evt["timestamp"] = time.Now().Format(time.RFC3339)
	evt["event_type"] = event
	evt["remote_addr"] = a.findIP(r)

	for k, v := range extraFields {
		evt[k] = v
	}

	headers := map[string]string{}
	for _, k := range a.Headers {
		if v := r.Header.Get(k); v != "" {
			headers[k] = v
		}
	}

	evt["headers"] = headers

	// Submit event to all specified targets
	for _, target := range a.Targets {
		if err := a.submitLog(target, evt); err != nil {
			return errors.Wrapf(err, "Could not submit log to target %q", target)
		}
	}

	return nil
}

func (a *auditLogger) findIP(r *http.Request) string {
	remoteAddr := strings.SplitN(r.RemoteAddr, ":", 2)[0]

	for _, hdr := range a.TrustedIPHeaders {
		if value := r.Header.Get(hdr); value != "" {
			return strings.SplitN(value, ",", 2)[0]
		}
	}

	return remoteAddr
}

func (a *auditLogger) submitLog(target string, event map[string]interface{}) error {
	u, err := url.Parse(target)
	if err != nil {
		return errors.Wrap(err, "Unable to parse target")
	}

	switch u.Scheme {
	case "fd":
		return a.submitLogFileDescriptor(u.Host, event)
	case "file":
		return a.submitLogFile(u.Path, event)
	default:
		return errors.Errorf("Unsupported target scheme %q", u.Scheme)
	}
}

func (a *auditLogger) submitLogFile(filename string, event map[string]interface{}) error {
	if err := os.MkdirAll(path.Dir(filename), 0600); err != nil {
		return errors.Wrap(err, "Unable to create required paths")
	}

	f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return errors.Wrap(err, "Unable to open audit file")
	}
	defer f.Close()

	return json.NewEncoder(f).Encode(event)
}

func (a *auditLogger) submitLogFileDescriptor(descriptor string, event map[string]interface{}) error {
	var w io.Writer

	switch descriptor {
	case "stdout":
		w = os.Stdout
	case "stderr":
		w = os.Stderr
	default:
		return errors.Errorf("Unsupported file descriptor %q", descriptor)
	}

	return errors.Wrap(json.NewEncoder(w).Encode(event), "Unable to marshal event")
}
