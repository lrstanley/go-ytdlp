.DEFAULT_GOAL := generate

export YTDLP_VERSION := 2025.12.08

license:
	curl -sL https://liam.sh/-/gh/g/license-header.sh | bash -s

clean:
	rm -rf ./cmd/patch-ytdlp/tmp/${YTDLP_VERSION} ./cmd/patch-ytdlp/export-${YTDLP_VERSION}.json

fetch:
	cd ./cmd/codegen && go mod tidy
	cd ./cmd/gen-jsonschema && go mod tidy
	go mod tidy

up:
	cd ./cmd/codegen && go get -u -t ./... && go mod tidy
	cd ./cmd/gen-jsonschema && go get -u -t ./... && go mod tidy
	cd ./_examples && go get -u -t ./... && go mod tidy
	go get -u -t ./... && go mod tidy

commit: generate
	git add --all \
		Makefile \
		*.gen.go *.gen_test.go \
		optiondata/*.gen.go
	git commit -m "chore(codegen): generate updated cli bindings"

edit-patch: clean patch
	cd ./cmd/patch-ytdlp/tmp/${YTDLP_VERSION} && ${EDITOR} yt_dlp/options.py && git diff > ../../export-options.patch

patch:
	@# git diff --minimal -U1 > ../../export-options.patch
	./cmd/patch-ytdlp/run.sh ${YTDLP_VERSION}

test: fetch
	GORACE='exitcode=1 halt_on_error=1' go test -v -race -timeout 3m -count 3 ./...

generate: license fetch patch
	rm -rf \
		*.gen.go *.gen_test.go \
		optiondata/*.gen.go
	cd ./cmd/codegen && go run . ../patch-ytdlp/export-${YTDLP_VERSION}.json ../../
	gofmt -e -s -w .
	cd ./cmd/gen-jsonschema && go run . ../../optiondata/
	go vet .
	go test -v ./...
