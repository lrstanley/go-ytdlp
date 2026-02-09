module github.com/lrstanley/go-ytdlp

go 1.24.0

toolchain go1.24.4

require (
	github.com/ProtonMail/go-crypto v1.3.0
	github.com/ulikunitz/xz v0.5.15
)

require (
	github.com/cloudflare/circl v1.6.3 // indirect
	golang.org/x/crypto v0.47.0 // indirect
	golang.org/x/sys v0.41.0 // indirect
)

// Testing dependencies. Not pulled when "go get"ing.
require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/stretchr/testify v1.11.1
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
