BINARY=gray

VERSION?="0.3.0"
COMMIT=$(shell git rev-parse --short HEAD)
BRANCH=$(shell git rev-parse --abbrev-ref HEAD)

# Package that we will store the ldflag variables
PREFIX="github.com/kylec725/graytorrent/cmd"

# Setup the -ldflags option for go build here, interpolate the variable values
LDFLAGS = -ldflags "-X ${PREFIX}.version=${VERSION} -X ${PREFIX}.commit=${COMMIT} -X ${PREFIX}.branch=${BRANCH}"

build:
	go build ${LDFLAGS} -o ${BINARY}

clean:
	go clean
	rm ${BINARY}

.PHONY: build clean
