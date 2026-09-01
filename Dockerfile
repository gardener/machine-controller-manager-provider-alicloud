#############      builder                          #############
FROM --platform=$BUILDPLATFORM golang:1.26.2 AS builder

ARG TARGETOS
ARG TARGETARCH

WORKDIR /go/src/github.com/gardener/machine-controller-manager-provider-alicloud
COPY . .

RUN GOOS=$TARGETOS GOARCH=$TARGETARCH .ci/build

#############      machine-controller               #############
FROM gcr.io/distroless/static-debian12:nonroot AS machine-controller
WORKDIR /

COPY --from=builder /go/src/github.com/gardener/machine-controller-manager-provider-alicloud/bin/rel/machine-controller /machine-controller
ENTRYPOINT ["/machine-controller"]
