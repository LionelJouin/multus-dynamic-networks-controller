FROM quay.io/projectquay/golang:1.19 as builder
ENV GOPATH=/go
RUN mkdir -p $GOPATH/src/github.com/maiqueb/multus-dynamic-networks-controller
WORKDIR $GOPATH/src/github.com/maiqueb/multus-dynamic-networks-controller
COPY . .
RUN GOOS=linux CGO_ENABLED=0 go build -o /dynamic-networks-controller ./cmd/dynamic-networks-controller

FROM registry.access.redhat.com/ubi8/ubi-minimal
COPY --from=builder /dynamic-networks-controller /dynamic-networks-controller

ARG git_url=https://github.com/k8snetworkplumbingwg/multus-dynamic-networks-controller.git
ARG git_sha=NONE

LABEL multi.GIT_URL=${git_url} \
      multi.GIT_SHA=${git_sha}

ENV DYNAMIC_NETWORK_CONTROLLER_COMMIT_HASH="${git_sha}"
