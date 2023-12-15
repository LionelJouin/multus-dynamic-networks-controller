GO ?= go
OCI_BIN ?= docker

IMAGE_REGISTRY ?= localhost:5000/k8snetworkplumbingwg
IMAGE_NAME ?= multus-dynamic-networks-controller
IMAGE_TAG ?= latest-amd64-debug
NAMESPACE ?= kube-system

CONTAINERD_SOCKET_PATH ?= "/run/containerd/containerd.sock"
CRIO_SOCKET_PATH ?= "/run/crio/crio.sock"
MULTUS_SOCKET_PATH ?= "/run/multus/multus.sock"

GIT_SHA := $(shell git describe --no-match  --always --abbrev=40 --dirty)

.PHONY: manifests \
        vendor

all: build test

build:
	$(GO) build -o bin/dynamic-networks-controller cmd/dynamic-networks-controller/networks-controller.go

clean:
	rm -rf bin/ manifests/

img-build: build
	$(OCI_BIN) build -t ${IMAGE_REGISTRY}/${IMAGE_NAME}:${IMAGE_TAG} -f images/Dockerfile --build-arg git_sha=$(GIT_SHA) . ; \
	docker push ${IMAGE_REGISTRY}/${IMAGE_NAME}:${IMAGE_TAG}

manifests:
	MULTUS_SOCKET_PATH=${MULTUS_SOCKET_PATH} IMAGE_REGISTRY=${IMAGE_REGISTRY} IMAGE_TAG=${IMAGE_TAG} CRI_SOCKET_PATH=${CONTAINERD_SOCKET_PATH} NAMESPACE=${NAMESPACE} hack/generate_manifests.sh
	CRIO_RUNTIME="yes" MULTUS_SOCKET_PATH=${MULTUS_SOCKET_PATH} IMAGE_REGISTRY=${IMAGE_REGISTRY} IMAGE_TAG=${IMAGE_TAG} CRI_SOCKET_PATH=${CRIO_SOCKET_PATH} NAMESPACE=${NAMESPACE} hack/generate_manifests.sh

test:
	$(GO) test -v -timeout=5s -count=1 ./pkg/controller/...

ginkgo:
	ginkgo --repeat=100 -timeout=240s ./pkg/controller/...

e2e/test:
	ginkgo -vv ./e2e/...

vendor:
	$(GO) mod tidy
	$(GO) mod vendor
