.DEFAULT_GOAL := generate

export YTDLP_VERSION := 2023.07.06

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

patch:
	./cmd/patch-ytdlp/run.sh ${YTDLP_VERSION}

generate: license go-fetch patch
	rm -rf *.gen.go
	go run github.com/lrstanley/go-ytdlp/cmd/codegen ./cmd/patch-ytdlp/export-${YTDLP_VERSION}.json
	gofmt -e -s -w *.go
	go vet *.go
	go test -v ./...
