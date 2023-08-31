.DEFAULT_GOAL := generate

license:
	curl -sL https://liam.sh/-/gh/g/license-header.sh | bash -s

up: go-upgrade-deps
	@echo

go-fetch:
	go mod download
	go mod tidy

go-upgrade-deps:
	go get -u ./...
	go mod tidy

go-upgrade-deps-patch:
	go get -u=patch ./...
	go mod tidy

commit: generate
	git add --all *.gen.go
	git commit -m "chore(codegen): generate updated cli bindings"

generate: license go-fetch
	rm -rf *.gen.go
	go run github.com/lrstanley/go-ytdlp/cmd/codegen ./yt-dlp-example-export.json
	gofmt -e -s -w *.go
	go vet *.go
	go test -v ./...
