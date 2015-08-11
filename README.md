# gh-downloader

Download a GitHub release asset from a semver version template

## Installation

Easy! You can just download the binary from the command line:

* Linux

```shell
curl -sL https://github.com/upfluence/etcdenv/releases/download/v0.0.1/gh-downloader-linux-amd64 > etcdenv
```

* OSX

```shell
curl -sL https://github.com/upfluence//releases/download/v0.0.1/gh-downloader-darwin-amd64 > etcdenv
```

If you would prefer compile the binary (assuming buildtools and Go are
installed) :

```shell
git clone git@github.com:upfluence/gh-downloader.git
cd gh-downloader
go get github.com/tools/godep
GOPATH=`pwd`/Godeps/_workspace go build -o gh-downloader .
```

## Usage

### Options

```
Usage of gh-downloader:
  -a="": Asset name
  -gh-token="": GitHub API token
  -o="": Output location
  -repository="": Repository
  -s="": Scheme of the release
  -v=false: Print the Version and exit
  -version=false: Print the version and exit
```

### Example

Let's say you have a repositiory `upfluence/foo` and all the release are
tagged such as `bar-0.0.1`, `bar-0.1.0` and so on.

you can download the latest release with:

```shell
gh-downloader -a mybin -gh-token="GITHUB_TOKEN.." -o mybin  -repository upfluence/foo -s "bar-x.x.x"
```

you can also download the latest release with the major version 3

```shell
gh-downloader -a mybin -gh-token="GITHUB_TOKEN.." -o mybin  -repository upfluence/foo -s "bar-3.x.x"
```
