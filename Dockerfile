FROM --platform=$BUILDPLATFORM golang:1.23 AS builder
WORKDIR /go/src/github.com/kubernetes-sigs/hyperv-csi-driver
RUN go env -w GOCACHE=/gocache GOMODCACHE=/gomodcache
COPY go.* .
ARG GOPROXY
RUN --mount=type=cache,target=/gomodcache go mod download
COPY . .
ARG TARGETOS
ARG TARGETARCH
ARG VERSION
ARG GOEXPERIMENT
RUN --mount=type=cache,target=/gomodcache --mount=type=cache,target=/gocache OS=$TARGETOS ARCH=$TARGETARCH make

FROM debian:bookworm-slim AS debian
COPY --from=builder /go/src/github.com/kubernetes-sigs/hyperv-csi-driver/bin/hyperv-csi-driver /bin/hyperv-csi-driver
RUN apt-get update
RUN apt-get install -y lsscsi
RUN groupadd -g 1000 app
RUN useradd -ms /bin/bash -u 1000 -g 1000 app
USER app
ENTRYPOINT ["/bin/hyperv-csi-driver"]
