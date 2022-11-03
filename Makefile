export CGO_ENABLED:=0

VERSION=$(shell git describe --tags --match=v* --always --dirty)

REPO=github.com/poseidon/scuttle
LOCAL_REPO?=poseidon/scuttle
IMAGE_REPO?=quay.io/poseidon/scuttle

LD_FLAGS="-w -X main.version=$(VERSION)"

.PHONY: all
all: build test vet fmt

.PHONY: build
build: bin/scuttle

.PHONY: bin/scuttle
bin/scuttle:
	@go build -o bin/scuttle -ldflags $(LD_FLAGS) $(REPO)/cmd/scuttle

run:
	./bin/scuttle -uncordon=true -drain=true -delete=true

.PHONY: test
test:
	@go test ./... -cover

.PHONY: vet
vet:
	@go vet -all ./...

.PHONY: fmt
fmt:
	@test -z $$(go fmt ./...)

image: \
	image-amd64 \
	image-arm64

image-%:
	buildah bud -f Dockerfile \
		-t $(LOCAL_REPO):$(VERSION)-$* \
		--arch $* --override-arch $* \
		--format=docker .
