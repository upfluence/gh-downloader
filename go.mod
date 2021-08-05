module github.com/upfluence/gh-downloader

go 1.16

require (
	github.com/google/go-github v13.0.1-0.20171030212440-724ae38c5030+incompatible
	github.com/google/go-querystring v1.0.0 // indirect
	github.com/upfluence/cfg v0.2.3
	golang.org/x/oauth2 v0.0.0-20180821212333-d2e6202438be
)

replace github.com/coreos/bbolt v1.3.4 => go.etcd.io/bbolt v1.3.4
