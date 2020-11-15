GIT_COMMIT := $(shell git rev-list -1 HEAD)
BUILD_DATE := $(shell date --iso-8601=seconds)

build:
	go build -ldflags "-X main.version=$(GIT_COMMIT) -X main.date=$(BUILD_DATE)" -o ~/.local/bin/co