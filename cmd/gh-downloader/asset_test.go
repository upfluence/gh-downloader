package main

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type dirEntry struct {
	Mode    os.FileMode
	Content string
}

func TestWriteAsset(t *testing.T) {
	var (
		errRead = errors.New("connection reset")

		mismatch = checksumVerifier(sumOf("other"))
	)

	for _, tt := range []struct {
		name         string
		haveReader   io.Reader
		haveOutput   string
		haveVerifier verifier
		wantErr      error
		wantStdout   string
		wantFiles    map[string]dirEntry
	}{
		{
			name:         "write to the output file",
			haveReader:   strings.NewReader(assetContent),
			haveOutput:   "tool",
			haveVerifier: noopVerifier,
			wantFiles: map[string]dirEntry{
				"tool": {Mode: 0o640, Content: assetContent},
			},
		},
		{
			name:         "write to stdout",
			haveReader:   strings.NewReader(assetContent),
			haveVerifier: noopVerifier,
			wantStdout:   assetContent,
			wantFiles:    map[string]dirEntry{},
		},
		{
			name:         "checksum mismatch leaves no file",
			haveReader:   strings.NewReader(assetContent),
			haveOutput:   "asset",
			haveVerifier: mismatch,
			wantErr:      errChecksumMismatch,
			wantFiles:    map[string]dirEntry{},
		},
		{
			name:         "checksum mismatch writes nothing to stdout",
			haveReader:   strings.NewReader(assetContent),
			haveVerifier: mismatch,
			wantErr:      errChecksumMismatch,
			wantFiles:    map[string]dirEntry{},
		},
		{
			name:         "failed download leaves no file",
			haveReader:   io.MultiReader(strings.NewReader("bin"), iotest.ErrReader(errRead)),
			haveOutput:   "asset",
			haveVerifier: noopVerifier,
			wantErr:      errRead,
			wantFiles:    map[string]dirEntry{},
		},
		{
			name:         "failed download writes nothing to stdout",
			haveReader:   io.MultiReader(strings.NewReader("bin"), iotest.ErrReader(errRead)),
			haveVerifier: noopVerifier,
			wantErr:      errRead,
			wantFiles:    map[string]dirEntry{},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var (
				dir = t.TempDir()

				output string
				stdout bytes.Buffer
			)

			t.Setenv("TMPDIR", dir)

			if tt.haveOutput != "" {
				output = filepath.Join(dir, tt.haveOutput)
			}

			err := writeAsset(
				tt.haveReader,
				output,
				0o640,
				&stdout,
				tt.haveVerifier,
			)

			require.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.wantStdout, stdout.String())
			assert.Equal(t, tt.wantFiles, readDir(t, dir))
		})
	}
}

func readDir(t *testing.T, dir string) map[string]dirEntry {
	t.Helper()

	des, err := os.ReadDir(dir)

	require.NoError(t, err)

	res := make(map[string]dirEntry, len(des))

	for _, de := range des {
		fi, err := de.Info()

		require.NoError(t, err)

		buf, err := os.ReadFile(filepath.Join(dir, de.Name()))

		require.NoError(t, err)

		res[de.Name()] = dirEntry{Mode: fi.Mode(), Content: string(buf)}
	}

	return res
}
