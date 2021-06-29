# gh-downloader

Download a GitHub release asset from a semver version template

## Installation

Easy! You can just download the binary from the command line:

* Linux

```shell
curl -sL https://github.com/upfluence/gh-downloader/releases/download/v0.0.1/gh-downloader-linux-amd64 > gh-downloader
```

* OSX

```shell
curl -sL https://github.com/upfluence/gh-downloader/releases/download/v0.0.1/gh-downloader-darwin-amd64 > gh-downloader
```

If you prefer compiling the binary (assuming buildtools and Go are
installed):

```shell
go install github.com/upfluence/gh-downloader/cmd/gh-downloader
```

## Usage

### Options

```
usage: gh-downloader [--gh-token] [--repository, -r] [--asset, -a] [--scheme, -s] [--latest] [--output, -o] [--mode]
Arguments:
	- GithubToken: string GitHub API token (env: GITHUB_TOKEN, UPF_GITHUB_TOKEN, flag: --gh-token)
	- Repository: main.repoConfig Repository (env: REPOSITORY, flag: --repository, -r)
	- Asset: string Asset name (env: ASSET, flag: --asset, -a)
	- Scheme: string Scheme of the release (env: SCHEME, flag: --scheme, -s)
	- Output: string Output location on disk, if left empty the file will be written on stdout (env: OUTPUT, flag: --output, -o)
	- FileMode: main.fileMode File mode to create the output file in (default: 0644) (env: FILEMODE, flag: --mode)
```

### Example

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
