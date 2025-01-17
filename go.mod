module github.com/lrstanley/go-ytdlp

go 1.22.0

toolchain go1.23.1

require github.com/ProtonMail/go-crypto v1.1.3

require (
	github.com/cloudflare/circl v1.5.0 // indirect
	golang.org/x/crypto v0.31.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
)

// Testing dependencies. Not pulled when "go get"ing.
require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/stretchr/testify v1.10.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
