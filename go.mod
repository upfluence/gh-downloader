module github.com/upfluence/gh-downloader

go 1.26.0

require (
	github.com/google/go-github/v93 v93.0.0
	github.com/stretchr/testify v1.11.1
	github.com/upfluence/cfg v0.2.3
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/google/go-querystring v1.2.0 // indirect
	github.com/pkg/errors v0.8.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/coreos/bbolt v1.3.4 => go.etcd.io/bbolt v1.3.4
