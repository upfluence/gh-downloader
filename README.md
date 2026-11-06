# gh-downloader

Download a GitHub release asset from a semver version template

## Installation

Easy! You can just download the binary from the command line:

* Linux

```shell
curl -sL https://github.com/upfluence/gh-downloader/releases/download/v0.1.0/gh-downloader-linux-amd64 > gh-downloader
```

* OSX

```shell
curl -sL https://github.com/upfluence/gh-downloader/releases/download/v0.1.0/gh-downloader-darwin-amd64 > gh-downloader
```

If you prefer compiling the binary (assuming buildtools and Go are
installed):

```shell
go install github.com/upfluence/gh-downloader/cmd/gh-downloader
```

## Usage

### CLI

#### Options

```
usage: gh-downloader [--gh-token] [--repository, -r] [--asset, -a] [--scheme, -s] [--output, -o] [--mode] [--if-not-exist] [--verify.mode] [--verify.checksum.sum]
Arguments:
	- GithubToken: string GitHub API token (env: GITHUB_TOKEN, UPF_GITHUB_TOKEN, flag: --gh-token)
	- Repository: main.repoConfig Repository (env: REPOSITORY, flag: --repository, -r)
	- Asset: main.templateConfig Asset name, it can be given as a go template with variable including: .tag, .repo, .arch, .os, .version (default: {{ .repo }}-{{ .os }}-{{ .arch }}-{{ .version }}) (env: ASSET, flag: --asset, -a)
	- Scheme: string Scheme of the release, if empty the latest release will be pulled (env: SCHEME, flag: --scheme, -s)
	- Output: string Output location on disk, if left empty the file will be written on stdout (env: OUTPUT, flag: --output, -o)
	- FileMode: main.fileMode File mode to create the output file in (default: 0644) (env: FILEMODE, flag: --mode)
	- IfNotExist: bool (env: IF_NOT_EXIST, flag: --if-not-exist)
	- Verify.Mode: main.verifyMode How to verify the asset before writing it: none, checksum (env: VERIFY_MODE, flag: --verify.mode)
	- Verify.Checksum.Sum: main.sha256Sum Expected SHA-256 of the asset, hex encoded with an optional sha256: prefix (env: VERIFY_CHECKSUM_SUM, flag: --verify.checksum.sum)
```

#### Example

Let's say you have a repositiory `upfluence/foo` and all the release are
tagged such as `bar-0.0.1`, `bar-0.1.0` and so on.

you can download the latest release with:

```shell
gh-downloader -a mybin -gh-token="GITHUB_TOKEN.." -o mybin  -repository upfluence/foo -s "bar-vx.x.x"
```

you can also download the latest release with the major version 3

```shell
gh-downloader -a mybin -gh-token="GITHUB_TOKEN.." -o mybin  -repository upfluence/foo -s "bar-v3.x.x"
```

#### Verifying the asset

With `--verify.mode checksum`, the asset is only written once its SHA-256
matches `--verify.checksum.sum`. On a mismatch, gh-downloader fails and
leaves neither the output file nor anything on stdout. Pin a release with
`-s` when using a checksum, otherwise the next release fails the check.

```shell
gh-downloader -a mybin -o mybin -repository upfluence/foo -s "bar-v1.2.3" \
  --verify.mode checksum \
  --verify.checksum.sum sha256:9601d410de178c829406b3099f488314d75474559c9810d5681d069791939ad2
```

Both can also be given through the `VERIFY_MODE` and `VERIFY_CHECKSUM_SUM`
environment variables.

### GitHub Action

```
on: push
jobs:
  gh-downloader:
    runs-on: ubuntu-20.04
    steps:
    - uses: upfluence/gh-downloader@master
      with:
        repo: upfluence/ubuild
        github-token: ${{ secrets.GITHUB_TOKEN }}
        asset: ubuild-linux-amd64
        output: ~/gh-action
```
