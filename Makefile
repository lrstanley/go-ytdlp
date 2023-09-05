.DEFAULT_GOAL := generate

export YTDLP_VERSION := 2023.07.06

license:
	curl -sL https://liam.sh/-/gh/g/license-header.sh | bash -s

up: go-upgrade-deps
	@echo

go-fetch:
	cd ./cmd/codegen && go mod download && go mod tidy
	go mod download && go mod tidy

go-upgrade-deps:
	cd ./cmd/codegen && go get -u ./... && go mod tidy
	go get -u ./... && go mod tidy

go-upgrade-deps-patch:
	cd ./cmd/patch-ytdlp && go get -u=patch ./... && go mod tidy
	go get -u=patch ./... && go mod tidy

commit: generate
	git add --all *.gen.go
	git commit -m "chore(codegen): generate updated cli bindings"

patch:
	./cmd/patch-ytdlp/run.sh ${YTDLP_VERSION}

generate: license go-fetch patch
	rm -rf *.gen.go
	cd ./cmd/codegen && go run . ../patch-ytdlp/export-${YTDLP_VERSION}.json ../../
	gofmt -e -s -w *.go
	go vet *.go
	go test -v ./...
