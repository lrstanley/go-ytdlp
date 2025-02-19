.DEFAULT_GOAL := generate

export YTDLP_VERSION := 2025.02.19

license:
	curl -sL https://liam.sh/-/gh/g/license-header.sh | bash -s

up: go-upgrade-deps
	@echo

clean:
	rm -rf ./cmd/patch-ytdlp/tmp/${YTDLP_VERSION} ./cmd/patch-ytdlp/export-${YTDLP_VERSION}.json

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
	git add --all \
		Makefile \
		*.gen.go *.gen_test.go \
		optiondata/*.gen.go
	git commit -m "chore(codegen): generate updated cli bindings"

edit-patch: clean patch
	cd ./cmd/patch-ytdlp/tmp/${YTDLP_VERSION} && ${EDITOR} yt_dlp/options.py && git diff > ../../export-options.patch

patch:
	./cmd/patch-ytdlp/run.sh ${YTDLP_VERSION}

generate: license go-fetch patch
	rm -rf \
		*.gen.go *.gen_test.go \
		optiondata/*.gen.go
	cd ./cmd/codegen && go run . ../patch-ytdlp/export-${YTDLP_VERSION}.json ../../
	gofmt -e -s -w .
	go vet .
	go test -v ./...
