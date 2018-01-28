package github

import (
	"bufio"
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strings"
	"text/template"
	"time"

	update "github.com/inconshreveable/go-update"
)

const (
	defaultTimeout      = 60 * time.Second
	defaultNamingScheme = `{{.ProductName}}_{{.GOOS}}_{{.GOARCH}}{{.EXT}}`
)

var (
	errReleaseNotFound = errors.New("Release not found")
)

// Updater is the core struct of the update library holding all configurations
type Updater struct {
	repo      string
	myVersion string

	HTTPClient     *http.Client
	RequestTimeout time.Duration
	Context        context.Context
	Filename       string

	releaseCache string
}

// NewUpdater initializes a new Updater and tries to guess the Filename
func NewUpdater(repo, myVersion string) (*Updater, error) {
	var err error
	u := &Updater{
		repo:      repo,
		myVersion: myVersion,

		HTTPClient:     http.DefaultClient,
		RequestTimeout: defaultTimeout,
		Context:        context.Background(),
	}

	u.Filename, err = u.compileFilename()

	return u, err
}

// HasUpdate checks which tag was used in the latest version and compares it to the current version. If it differs the function will return true. No comparison is done to determine whether the found version is higher than the current one.
func (u *Updater) HasUpdate(forceRefresh bool) (bool, error) {
	if forceRefresh {
		u.releaseCache = ""
	}

	latest, err := u.getLatestRelease()
	switch err {
	case nil:
		return u.myVersion != latest, nil
	case errReleaseNotFound:
		return false, nil
	default:
		return false, err
	}
}

// Apply downloads the new binary from Github, fetches the SHA256 sum from the SHA256SUMS file and applies the update to the currently running binary
func (u *Updater) Apply() error {
	updateAvailable, err := u.HasUpdate(false)
	if err != nil {
		return err
	}
	if !updateAvailable {
		return nil
	}

	checksum, err := u.getSHA256(u.Filename)
	if err != nil {
		return err
	}

	newRelease, err := u.getFile(u.Filename)
	if err != nil {
		return err
	}
	defer newRelease.Close()

	return update.Apply(newRelease, update.Options{
		Checksum: checksum,
	})
}

func (u Updater) getSHA256(filename string) ([]byte, error) {
	shaFile, err := u.getFile("SHA256SUMS")
	if err != nil {
		return nil, err
	}
	defer shaFile.Close()

	scanner := bufio.NewScanner(shaFile)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.Contains(line, u.Filename) {
			continue
		}

		return hex.DecodeString(line[0:64])
	}

	return nil, fmt.Errorf("No SHA256 found for file %q", u.Filename)
}

func (u Updater) getFile(filename string) (io.ReadCloser, error) {
	release, err := u.getLatestRelease()
	if err != nil {
		return nil, err
	}

	requestURL := fmt.Sprintf("https://github.com/%s/releases/download/%s/%s", u.repo, release, filename)
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return nil, err
	}

	ctx, _ := context.WithTimeout(u.Context, u.RequestTimeout)

	res, err := u.HTTPClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("File not found: %q", requestURL)
	}

	return res.Body, nil
}

func (u *Updater) getLatestRelease() (string, error) {
	if u.releaseCache != "" {
		return u.releaseCache, nil
	}

	result := struct {
		TagName string `json:"tag_name"`
	}{}

	requestURL := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", u.repo)
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return "", err
	}

	ctx, cancel := context.WithTimeout(u.Context, u.RequestTimeout)
	defer cancel()

	res, err := u.HTTPClient.Do(req.WithContext(ctx))
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if err = json.NewDecoder(res.Body).Decode(&result); err != nil {
		return "", err
	}

	if res.StatusCode != 200 || result.TagName == "" {
		return "", errReleaseNotFound
	}

	u.releaseCache = result.TagName

	return result.TagName, nil
}

func (u Updater) compileFilename() (string, error) {
	repoName := strings.Split(u.repo, "/")
	if len(repoName) != 2 {
		return "", errors.New("Repository name not in format <owner>/<repository>")
	}

	tpl, err := template.New("filename").Parse(defaultNamingScheme)
	if err != nil {
		return "", err
	}

	var ext string
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}

	buf := bytes.NewBuffer([]byte{})
	if err = tpl.Execute(buf, map[string]interface{}{
		"GOOS":        runtime.GOOS,
		"GOARCH":      runtime.GOARCH,
		"EXT":         ext,
		"ProductName": repoName[1],
	}); err != nil {
		return "", err
	}

	return buf.String(), nil
}
