package update

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"strings"
	"testing"

	"github.com/apiservicesac/dooprint/internal/testutil"
)

func TestNewer(t *testing.T) {
	cases := []struct {
		candidate string
		current   string
		expected  bool
	}{
		{"1.0.1", "1.0.0", true},
		{"1.1.0", "1.0.9", true},
		{"2.0.0", "1.9.9", true},
		{"1.0.0", "1.0.0", false},
		{"1.0.0", "1.0.1", false},
		{"1.0.10", "1.0.9", true},
		{"1.0.9", "1.0.10", false},
		// Anything unreadable is never an update.
		{"", "1.0.0", false},
		{"latest", "1.0.0", false},
		{"1.0.0-rc1", "1.0.0", false},
		{"1.0", "1.0.0", false},
	}
	for _, c := range cases {
		testutil.ExpectedEqual(t, newer(c.candidate, c.current), c.expected,
			fmt.Sprintf("newer(%q, %q)", c.candidate, c.current))
	}
}

func TestAssetName(t *testing.T) {
	name, ok := assetName("1.2.3")
	testutil.ExpectedTrue(t, ok, "expected a published build for the test system")
	testutil.ExpectedTrue(t, name != "", "expected a release file name")
	testutil.ExpectedTrue(t, strings.Contains(name, "1.2.3"),
		fmt.Sprintf("expected the version in %q", name))
}

func TestBinaryInPlainFile(t *testing.T) {
	binary, err := binaryIn("dooprint-1.2.3-windows-amd64.exe", []byte("binary"))
	testutil.ExpectedNoError(t, err)
	testutil.ExpectedEqual(t, string(binary), "binary")
}

func TestBinaryInArchive(t *testing.T) {
	archive := tarball(t, map[string]string{
		"dooprint/install.sh": "#!/bin/sh",
		"dooprint/dooprint":   "binary",
	})

	binary, err := binaryIn("dooprint-1.2.3-linux-amd64.tar.gz", archive)
	testutil.ExpectedNoError(t, err)
	testutil.ExpectedEqual(t, string(binary), "binary")
}

func TestBinaryInArchiveWithoutIt(t *testing.T) {
	archive := tarball(t, map[string]string{"dooprint/install.sh": "#!/bin/sh"})

	_, err := binaryIn("dooprint-1.2.3-linux-amd64.tar.gz", archive)
	testutil.ExpectedError(t, err)
}

// tarball builds a tar.gz with the given files, like the published Linux release.
func tarball(t *testing.T, files map[string]string) []byte {
	t.Helper()
	buffer := &bytes.Buffer{}
	gz := gzip.NewWriter(buffer)
	archive := tar.NewWriter(gz)
	for name, content := range files {
		header := &tar.Header{Name: name, Mode: 0o755, Size: int64(len(content)), Typeflag: tar.TypeReg}
		testutil.ExpectedNoError(t, archive.WriteHeader(header))
		_, err := archive.Write([]byte(content))
		testutil.ExpectedNoError(t, err)
	}
	testutil.ExpectedNoError(t, archive.Close())
	testutil.ExpectedNoError(t, gz.Close())
	return buffer.Bytes()
}
