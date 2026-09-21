// Package update looks for published releases of dooprint and installs one over the running
// binary.
//
// Nothing happens on its own: the device asks GitHub for the latest release when someone presses
// the button in the web interface, and downloads it only when the update is confirmed. The new
// binary takes the place of the running one and the process exits, which is how it restarts:
// systemd on Linux and the service manager on Windows start it again.
package update

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/apiservicesac/dooprint/internal/logger"
)

const (
	// Repo is where the releases are published.
	Repo = "apiservicesac/dooprint"

	latestURL       = "https://api.github.com/repos/" + Repo + "/releases/latest"
	downloadURL     = "https://github.com/" + Repo + "/releases/download"
	checkTimeout    = 15 * time.Second
	downloadTimeout = 10 * time.Minute
	// maxDownload caps what is read from the release, well above the size of the binary.
	maxDownload = 256 << 20
	// oldSuffix is the name the replaced binary is moved to on Windows, where a running
	// executable cannot be overwritten but can be renamed.
	oldSuffix = ".old"
)

// Status is what the web interface shows about updates.
type Status struct {
	Current   string `json:"current"`
	Latest    string `json:"latest,omitempty"`
	Available bool   `json:"available"`
	Notes     string `json:"notes,omitempty"`
	CheckedAt string `json:"checkedAt,omitempty"`
	Error     string `json:"error,omitempty"`
	// Supported is false on a system with no published build, where only the version shows.
	Supported bool `json:"supported"`
}

// Checker keeps the result of the last check, so the interface can show it again without
// asking GitHub every time it is opened.
type Checker struct {
	current string
	client  *http.Client
	mu      sync.Mutex
	status  Status
}

func New(current string) *Checker {
	return &Checker{
		current: current,
		client:  &http.Client{Timeout: checkTimeout},
		status:  Status{Current: current, Supported: supported()},
	}
}

// Status is the result of the last check, or just the running version before the first one.
func (c *Checker) Status() Status {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.status
}

// Check asks GitHub for the latest release. It downloads nothing.
func (c *Checker) Check() Status {
	status := Status{Current: c.current, Supported: supported(), CheckedAt: time.Now().Format(time.RFC3339)}

	latest, notes, err := c.latest()
	if err != nil {
		status.Error = err.Error()
		logger.Warnf("Update check failed: %v", err)
	} else {
		status.Latest = latest
		status.Notes = notes
		status.Available = newer(latest, c.current)
	}

	c.mu.Lock()
	c.status = status
	c.mu.Unlock()
	return status
}

// Install downloads the latest release and puts it in place of the running binary. The caller
// restarts the service afterwards: the new binary only runs from the next start.
func (c *Checker) Install() error {
	if !supported() {
		return fmt.Errorf("there is no published build for %s/%s", runtime.GOOS, runtime.GOARCH)
	}
	status := c.Status()
	if status.Latest == "" {
		status = c.Check()
		if status.Error != "" {
			return fmt.Errorf("%s", status.Error)
		}
	}
	if !status.Available {
		return fmt.Errorf("version %s is already the latest one", c.current)
	}

	asset, _ := assetName(status.Latest)
	logger.Infof("Update: downloading %s", asset)
	body, err := c.download(status.Latest, asset)
	if err != nil {
		return err
	}
	if err := c.verify(status.Latest, asset, body); err != nil {
		return err
	}
	binary, err := extract(asset, body)
	if err != nil {
		return err
	}
	if err := replaceRunning(binary); err != nil {
		return err
	}
	logger.Infof("Update: version %s installed, restarting", status.Latest)
	return nil
}

// latest returns the version of the latest release and its notes.
func (c *Checker) latest() (string, string, error) {
	var release struct {
		TagName string `json:"tag_name"`
		Body    string `json:"body"`
		Draft   bool   `json:"draft"`
	}
	if err := c.getJSON(latestURL, &release); err != nil {
		return "", "", err
	}
	version := strings.TrimPrefix(strings.TrimSpace(release.TagName), "v")
	if version == "" {
		return "", "", fmt.Errorf("the latest release has no version")
	}
	return version, strings.TrimSpace(release.Body), nil
}

func (c *Checker) getJSON(url string, into any) error {
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("User-Agent", "dooprint/"+c.current)

	response, err := c.client.Do(request)
	if err != nil {
		return fmt.Errorf("cannot reach GitHub: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("GitHub answered %s", response.Status)
	}
	return json.NewDecoder(io.LimitReader(response.Body, maxDownload)).Decode(into)
}

// download reads a file of the release into memory.
func (c *Checker) download(version, name string) ([]byte, error) {
	client := &http.Client{Timeout: downloadTimeout}
	url := fmt.Sprintf("%s/v%s/%s", downloadURL, version, name)
	response, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("cannot download %s: %w", name, err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cannot download %s: GitHub answered %s", name, response.Status)
	}
	return io.ReadAll(io.LimitReader(response.Body, maxDownload))
}

// verify checks the downloaded file against the SHA256SUMS of the release.
func (c *Checker) verify(version, name string, body []byte) error {
	sums, err := c.download(version, "SHA256SUMS")
	if err != nil {
		return err
	}
	sum := sha256.Sum256(body)
	want := ""
	for line := range strings.Lines(string(sums)) {
		fields := strings.Fields(line)
		if len(fields) == 2 && strings.TrimPrefix(fields[1], "*") == name {
			want = fields[0]
		}
	}
	if want == "" {
		return fmt.Errorf("%s is not listed in SHA256SUMS", name)
	}
	if want != hex.EncodeToString(sum[:]) {
		return fmt.Errorf("the downloaded file does not match its checksum")
	}
	return nil
}

// extract returns the binary inside the downloaded file: the Linux release is a tar.gz with the
// binary and its installer, the Windows one is the executable itself.
func extract(name string, body []byte) ([]byte, error) {
	if !strings.HasSuffix(name, ".tar.gz") {
		return body, nil
	}
	gz, err := gzip.NewReader(bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("the downloaded file is not readable: %w", err)
	}
	defer gz.Close()

	archive := tar.NewReader(gz)
	for {
		header, err := archive.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("the downloaded file is not readable: %w", err)
		}
		if header.Typeflag == tar.TypeReg && filepath.Base(header.Name) == "dooprint" {
			return io.ReadAll(io.LimitReader(archive, maxDownload))
		}
	}
	return nil, fmt.Errorf("the downloaded file has no dooprint binary")
}

// replaceRunning writes the binary next to the running one and takes its place. On Linux the
// file is replaced directly; on Windows the running executable is moved aside first, because it
// cannot be overwritten while it runs.
func replaceRunning(binary []byte) error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot find the running program: %w", err)
	}
	if resolved, err := filepath.EvalSymlinks(exe); err == nil {
		exe = resolved
	}

	next := filepath.Join(filepath.Dir(exe), ".dooprint-update")
	if err := os.WriteFile(next, binary, 0o755); err != nil {
		return fmt.Errorf("cannot write the new version next to %s: %w", exe, err)
	}
	if runtime.GOOS == "windows" {
		old := exe + oldSuffix
		os.Remove(old)
		if err := os.Rename(exe, old); err != nil {
			os.Remove(next)
			return fmt.Errorf("cannot move the running program aside: %w", err)
		}
	}
	if err := os.Rename(next, exe); err != nil {
		os.Remove(next)
		return fmt.Errorf("cannot put the new version in place: %w", err)
	}
	return nil
}

// CleanOld removes the replaced binary left by a Windows update. It is called on start, when
// the old executable is no longer running.
func CleanOld() {
	exe, err := os.Executable()
	if err != nil {
		return
	}
	os.Remove(exe + oldSuffix)
}

// supported is true on a system with a published build.
func supported() bool {
	_, ok := assetName("0")
	return ok
}

// assetName is the release file for this system.
func assetName(version string) (string, bool) {
	if runtime.GOARCH != "amd64" {
		return "", false
	}
	switch runtime.GOOS {
	case "linux":
		return fmt.Sprintf("dooprint-%s-linux-amd64.tar.gz", version), true
	case "windows":
		return fmt.Sprintf("dooprint-%s-windows-amd64.exe", version), true
	}
	return "", false
}

// newer compares two versions like 1.2.10, where a version with a suffix such as 1.2.0-rc1 is
// older than the plain one. An unreadable version is never newer.
func newer(candidate, current string) bool {
	left, leftSuffix, leftOk := parse(candidate)
	right, rightSuffix, rightOk := parse(current)
	if !leftOk || !rightOk {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return left[i] > right[i]
		}
	}
	return rightSuffix != "" && leftSuffix == ""
}

// parse splits a version into its three numbers and its suffix.
func parse(version string) ([3]int, string, bool) {
	version = strings.TrimPrefix(strings.TrimSpace(version), "v")
	suffix := ""
	if dash := strings.IndexAny(version, "-+"); dash >= 0 {
		version, suffix = version[:dash], version[dash+1:]
	}
	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		return [3]int{}, "", false
	}
	numbers := [3]int{}
	for i, part := range parts {
		number, err := strconv.Atoi(part)
		if err != nil {
			return [3]int{}, "", false
		}
		numbers[i] = number
	}
	return numbers, suffix, true
}
