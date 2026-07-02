export CGO_ENABLED:=0

VERSION=$(shell git describe --tags --match=v* --always --dirty)

REPO=github.com/poseidon/scuttle
LOCAL_REPO?=poseidon/scuttle
IMAGE_REPO?=quay.io/poseidon/scuttle

PLATFORM_amd64=linux/amd64
PLATFORM_arm64=linux/arm64/v8

LD_FLAGS="-w -X main.version=$(VERSION)"

.PHONY: all
all: build test vet fmt

.PHONY: build
build:
	@go build -o bin/scuttle -ldflags $(LD_FLAGS) $(REPO)/cmd/scuttle

.PHONY: test
test:
	@go test ./... -cover

.PHONY: vet
vet:
	@go vet -all ./...

.PHONY: fmt
fmt:
	@test -z $$(go fmt ./...)

.PHONY: lint
lint:
	@golangci-lint run ./...

image: \
	image-amd64 \
	image-arm64

image-%:
	podman build -f Dockerfile \
		-t $(LOCAL_REPO):$(VERSION)-$* \
		--platform $(PLATFORM_$*) .

push: \
	push-amd64 \
	push-arm64

push-%:
	podman tag $(LOCAL_REPO):$(VERSION)-$* $(IMAGE_REPO):$(VERSION)-$*
	podman push $(IMAGE_REPO):$(VERSION)-$*

manifest:
	podman manifest create $(IMAGE_REPO):$(VERSION)
	podman manifest add $(IMAGE_REPO):$(VERSION) docker://$(IMAGE_REPO):$(VERSION)-amd64
	podman manifest add $(IMAGE_REPO):$(VERSION) docker://$(IMAGE_REPO):$(VERSION)-arm64
	podman manifest inspect $(IMAGE_REPO):$(VERSION)
	podman manifest push $(IMAGE_REPO):$(VERSION) docker://$(IMAGE_REPO):$(VERSION)
