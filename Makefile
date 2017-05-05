ARGS=
VERBOSE=

.PHONY: build
build:
	go build -ldflags "-X main.appVersion=$$(git describe --tags)" ./cmd/clasy

.PHONY: debug
debug: build
	./clasy ${ARGS}

.PHONY: deps
deps:
	go get -v github.com/Masterminds/glide
	glide install
