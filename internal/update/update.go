// Package update installs a published release of dooprint over the running binary.
//
// Nothing happens on its own: the web interface asks for the latest release when someone presses
// the button, and downloads it only when the update is confirmed. The new binary takes the place
// of the running one and the process exits, which is how it restarts: systemd on Linux and the
// service manager on Windows start it again.
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
	"time"

	"github.com/apiservicesac/dooprint/internal/logger"
)

const (
	repo        = "apiservicesac/dooprint"
	latestURL   = "https://api.github.com/repos/" + repo + "/releases/latest"
	downloadURL = "https://github.com/" + repo + "/releases/download"
	// maxSize caps what is read from a release, well above the size of the binary.
	maxSize = 256 << 20
	// oldSuffix names the replaced binary on Windows, where a running executable cannot be
	// overwritten but can be renamed.
	oldSuffix = ".old"
)

var client = &http.Client{Timeout: 10 * time.Minute}

// Release is the latest published version, next to the running one.
type Release struct {
	Current string `json:"current"`
	Latest  string `json:"latest"`
	Newer   bool   `json:"newer"`
	Notes   string `json:"notes,omitempty"`
}

// Latest asks GitHub which version is published. It downloads nothing.
func Latest(current string) (Release, error) {
	var answer struct {
		TagName string `json:"tag_name"`
		Body    string `json:"body"`
	}
	if err := getJSON(latestURL, &answer); err != nil {
		return Release{}, err
	}
	latest := strings.TrimPrefix(strings.TrimSpace(answer.TagName), "v")
	if latest == "" {
		return Release{}, fmt.Errorf("the latest release has no version")
	}
	return Release{
		Current: current,
		Latest:  latest,
		Newer:   newer(latest, current),
		Notes:   strings.TrimSpace(answer.Body),
	}, nil
}

// Install puts the latest release in place of the running binary. The caller restarts the
// service afterwards: the new version only runs from the next start.
func Install(current string) (Release, error) {
	release, err := Latest(current)
	if err != nil {
		return Release{}, err
	}
	if !release.Newer {
		return release, fmt.Errorf("version %s is already the latest one", current)
	}
	name, ok := assetName(release.Latest)
	if !ok {
		return release, fmt.Errorf("there is no published build for %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	logger.Infof("Update: downloading %s", name)
	file, err := download(release.Latest, name)
	if err != nil {
		return release, err
	}
	if err := checkSum(release.Latest, name, file); err != nil {
		return release, err
	}
	binary, err := binaryIn(name, file)
	if err != nil {
		return release, err
	}
	if err := replaceRunning(binary); err != nil {
		return release, err
	}
	logger.Infof("Update: version %s installed", release.Latest)
	return release, nil
}

// CleanOld removes the binary a Windows update replaced. It runs on start, when the old
// executable is no longer running.
func CleanOld() {
	if exe, err := os.Executable(); err == nil {
		os.Remove(exe + oldSuffix)
	}
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

func getJSON(url string, into any) error {
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("User-Agent", "dooprint")

	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("cannot reach GitHub: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("GitHub answered %s", response.Status)
	}
	return json.NewDecoder(io.LimitReader(response.Body, maxSize)).Decode(into)
}

// download reads a file of a release into memory.
func download(version, name string) ([]byte, error) {
	response, err := client.Get(fmt.Sprintf("%s/v%s/%s", downloadURL, version, name))
	if err != nil {
		return nil, fmt.Errorf("cannot download %s: %w", name, err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cannot download %s: GitHub answered %s", name, response.Status)
	}
	return io.ReadAll(io.LimitReader(response.Body, maxSize))
}

// checkSum compares the downloaded file with the SHA256SUMS published with the release.
func checkSum(version, name string, file []byte) error {
	sums, err := download(version, "SHA256SUMS")
	if err != nil {
		return err
	}
	sum := sha256.Sum256(file)
	for line := range strings.Lines(string(sums)) {
		// Each line is "<sum>  <file>", with an asterisk before the name in binary mode.
		fields := strings.Fields(line)
		if len(fields) == 2 && strings.TrimPrefix(fields[1], "*") == name {
			if fields[0] != hex.EncodeToString(sum[:]) {
				return fmt.Errorf("the downloaded file does not match its checksum")
			}
			return nil
		}
	}
	return fmt.Errorf("%s is not listed in SHA256SUMS", name)
}

// binaryIn returns the program inside a downloaded file: the Linux release is a tar.gz with the
// binary and its installer, the Windows one is the executable itself.
func binaryIn(name string, file []byte) ([]byte, error) {
	if !strings.HasSuffix(name, ".tar.gz") {
		return file, nil
	}
	unreadable := func(err error) error {
		return fmt.Errorf("the downloaded file is not readable: %w", err)
	}

	gz, err := gzip.NewReader(bytes.NewReader(file))
	if err != nil {
		return nil, unreadable(err)
	}
	defer gz.Close()

	archive := tar.NewReader(gz)
	for {
		header, err := archive.Next()
		if err == io.EOF {
			return nil, fmt.Errorf("the downloaded file has no dooprint binary")
		}
		if err != nil {
			return nil, unreadable(err)
		}
		if header.Typeflag == tar.TypeReg && filepath.Base(header.Name) == "dooprint" {
			return io.ReadAll(io.LimitReader(archive, maxSize))
		}
	}
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
		os.Remove(exe + oldSuffix)
		if err := os.Rename(exe, exe+oldSuffix); err != nil {
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

// newer compares two versions like 1.2.10. A version that cannot be read is never newer, so a
// tag out of the usual shape never starts an update.
func newer(candidate, current string) bool {
	left, leftOk := numbers(candidate)
	right, rightOk := numbers(current)
	if !leftOk || !rightOk {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return left[i] > right[i]
		}
	}
	return false
}

// numbers splits "1.2.10" into its three numbers.
func numbers(version string) ([3]int, bool) {
	parts := strings.Split(strings.TrimSpace(version), ".")
	if len(parts) != 3 {
		return [3]int{}, false
	}
	result := [3]int{}
	for i, part := range parts {
		number, err := strconv.Atoi(part)
		if err != nil {
			return [3]int{}, false
		}
		result[i] = number
	}
	return result, true
}
