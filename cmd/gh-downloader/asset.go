package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/google/go-github/v93/github"
)

func downloadAsset(ctx context.Context, cl *releaseClient, asset *github.ReleaseAsset, output string, mode os.FileMode, stdout io.Writer) error {
	rc, err := cl.openAsset(ctx, asset)

	if err != nil {
		return err
	}

	defer rc.Close()

	return writeAsset(rc, output, mode, stdout)
}

// writeAsset stages the asset in a temporary file and only moves it to the
// output once fully downloaded, so a failed download never leaves a partial
// file behind for --if-not-exist to pick up.
func writeAsset(r io.Reader, output string, mode os.FileMode, stdout io.Writer) error {
	var dir string

	if output != "" {
		dir = filepath.Dir(output)
	}

	f, err := os.CreateTemp(dir, ".gh-downloader-*")

	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}

	defer func() {
		f.Close()
		os.Remove(f.Name())
	}()

	if _, err := io.Copy(f, r); err != nil {
		return fmt.Errorf("download asset: %w", err)
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("close asset: %w", err)
	}

	if output == "" {
		return copyFile(stdout, f.Name())
	}

	if err := os.Chmod(f.Name(), mode); err != nil {
		return fmt.Errorf("chmod asset: %w", err)
	}

	if err := os.Rename(f.Name(), output); err != nil {
		return fmt.Errorf("move asset to output: %w", err)
	}

	return nil
}

func copyFile(w io.Writer, path string) error {
	f, err := os.Open(path)

	if err != nil {
		return fmt.Errorf("open asset: %w", err)
	}

	defer f.Close()

	if _, err := io.Copy(w, f); err != nil {
		return fmt.Errorf("write asset: %w", err)
	}

	return nil
}
