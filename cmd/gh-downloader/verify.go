package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"io"
	"strings"
)

var (
	errChecksumMismatch = errors.New("checksum mismatch")
	errChecksumMissing  = errors.New("--verify.checksum.sum is required in checksum mode")
)

type verifyMode int

const (
	verifyModeNone verifyMode = iota
	verifyModeChecksum
)

var verifyModeNames = map[verifyMode]string{
	verifyModeNone:     "none",
	verifyModeChecksum: "checksum",
}

func (vm *verifyMode) Parse(v string) error {
	for m, n := range verifyModeNames {
		if n == v {
			*vm = m

			return nil
		}
	}

	return fmt.Errorf("invalid verify mode %q", v)
}

func (vm verifyMode) String() string { return verifyModeNames[vm] }

// sha256Sum is a lowercase hex encoded SHA-256 digest.
type sha256Sum string

func newSHA256Sum(digest []byte) sha256Sum {
	return sha256Sum(hex.EncodeToString(digest))
}

func (s *sha256Sum) Parse(v string) error {
	sum := strings.ToLower(strings.TrimPrefix(v, "sha256:"))
	buf, err := hex.DecodeString(sum)

	if err != nil {
		return fmt.Errorf("invalid sha256 %q: %w", v, err)
	}

	if len(buf) != sha256.Size {
		return fmt.Errorf("invalid sha256 %q: %d bytes long", v, len(buf))
	}

	*s = sha256Sum(sum)

	return nil
}

type checksumConfig struct {
	Sum sha256Sum `flag:"sum" help:"Expected SHA-256 of the asset, hex encoded with an optional sha256: prefix"`
}

type verifyConfig struct {
	Mode     verifyMode     `flag:"mode"     help:"How to verify the asset before writing it: none, checksum"`
	Checksum checksumConfig `flag:"checksum"`
}

// verifier wraps the writer the asset is downloaded into. Closing the
// returned writer closes the wrapped one and fails if the asset written so
// far does not pass the verification.
type verifier func(io.WriteCloser) io.WriteCloser

func noopVerifier(w io.WriteCloser) io.WriteCloser { return w }

func (vc verifyConfig) verifier() (verifier, error) {
	switch vc.Mode {
	case verifyModeNone:
		return noopVerifier, nil
	case verifyModeChecksum:
		if vc.Checksum.Sum == "" {
			return nil, errChecksumMissing
		}

		return checksumVerifier(vc.Checksum.Sum), nil
	}

	return nil, fmt.Errorf("unsupported verify mode %q", vc.Mode)
}

func checksumVerifier(want sha256Sum) verifier {
	return func(w io.WriteCloser) io.WriteCloser {
		return &checksumWriter{WriteCloser: w, h: sha256.New(), want: want}
	}
}

type checksumWriter struct {
	io.WriteCloser

	h    hash.Hash
	want sha256Sum
}

func (cw *checksumWriter) Write(p []byte) (int, error) {
	n, err := cw.WriteCloser.Write(p)

	cw.h.Write(p[:n])

	if err != nil {
		return n, fmt.Errorf("write: %w", err)
	}

	return n, nil
}

func (cw *checksumWriter) Close() error {
	if err := cw.WriteCloser.Close(); err != nil {
		return fmt.Errorf("close: %w", err)
	}

	if got := newSHA256Sum(cw.h.Sum(nil)); got != cw.want {
		return fmt.Errorf("%w: got %s, want %s", errChecksumMismatch, got, cw.want)
	}

	return nil
}
