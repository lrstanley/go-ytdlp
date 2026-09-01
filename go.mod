module github.com/lrstanley/go-ytdlp

go 1.25.0

toolchain go1.26.0

require (
	github.com/ProtonMail/go-crypto v1.4.1
	github.com/ulikunitz/xz v0.5.16
)

require (
	github.com/cloudflare/circl v1.6.5 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
	golang.org/x/crypto v0.55.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
)

// Testing dependencies. Not pulled when "go get"ing.
require github.com/stretchr/testify v1.12.1
