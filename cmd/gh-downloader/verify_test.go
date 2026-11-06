package main

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const emptySum = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"

func sumOf(v string) sha256Sum {
	digest := sha256.Sum256([]byte(v))

	return newSHA256Sum(digest[:])
}

func TestVerifyModeParse(t *testing.T) {
	for _, tt := range []struct {
		name      string
		haveValue string
		want      verifyMode
		assertErr assert.ErrorAssertionFunc
	}{
		{
			name:      "none",
			haveValue: "none",
			want:      verifyModeNone,
			assertErr: assert.NoError,
		},
		{
			name:      "checksum mode",
			haveValue: "checksum",
			want:      verifyModeChecksum,
			assertErr: assert.NoError,
		},
		{
			name:      "unknown mode",
			haveValue: "gh-attestation",
			assertErr: assert.Error,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var got verifyMode

			err := got.Parse(tt.haveValue)

			tt.assertErr(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSHA256SumParse(t *testing.T) {
	for _, tt := range []struct {
		name      string
		haveValue string
		want      sha256Sum
		assertErr assert.ErrorAssertionFunc
	}{
		{
			name:      "hex",
			haveValue: emptySum,
			want:      sumOf(""),
			assertErr: assert.NoError,
		},
		{
			name:      "with the sha256 prefix",
			haveValue: "sha256:" + emptySum,
			want:      sumOf(""),
			assertErr: assert.NoError,
		},
		{
			name:      "uppercase",
			haveValue: strings.ToUpper(emptySum),
			want:      sumOf(""),
			assertErr: assert.NoError,
		},
		{
			name:      "not hex",
			haveValue: "not-a-sum",
			assertErr: assert.Error,
		},
		{
			name:      "too short",
			haveValue: emptySum[:62],
			assertErr: assert.Error,
		},
		{
			name:      "empty",
			assertErr: assert.Error,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var got sha256Sum

			err := got.Parse(tt.haveValue)

			tt.assertErr(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

type recordWriter struct {
	bytes.Buffer

	closed   bool
	closeErr error
}

func (rw *recordWriter) Close() error {
	rw.closed = true

	return rw.closeErr
}

type recorded struct {
	Content string
	Closed  bool
}

func checksum(sum sha256Sum) verifyConfig {
	return verifyConfig{Mode: verifyModeChecksum, Checksum: checksumConfig{Sum: sum}}
}

func TestVerifier(t *testing.T) {
	errClose := errors.New("disk full")

	for _, tt := range []struct {
		name         string
		haveConfig   verifyConfig
		haveCloseErr error
		wantErr      error
		wantCloseErr error
	}{
		{
			name: "no verification",
		},
		{
			name:       "matching checksum",
			haveConfig: checksum(sumOf(assetContent)),
		},
		{
			name:         "checksum mismatch",
			haveConfig:   checksum(sumOf("other")),
			wantCloseErr: errChecksumMismatch,
		},
		{
			name:         "wrapped writer fails to close",
			haveConfig:   checksum(sumOf(assetContent)),
			haveCloseErr: errClose,
			wantCloseErr: errClose,
		},
		{
			name:       "checksum without a sum",
			haveConfig: verifyConfig{Mode: verifyModeChecksum},
			wantErr:    errChecksumMissing,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			v, err := tt.haveConfig.verifier()

			require.ErrorIs(t, err, tt.wantErr)

			if err != nil {
				return
			}

			rw := recordWriter{closeErr: tt.haveCloseErr}
			w := v(&rw)

			_, err = io.WriteString(w, assetContent)

			require.NoError(t, err)

			require.ErrorIs(t, w.Close(), tt.wantCloseErr)
			assert.Equal(
				t,
				recorded{Content: assetContent, Closed: true},
				recorded{Content: rw.String(), Closed: rw.closed},
			)
		})
	}
}
