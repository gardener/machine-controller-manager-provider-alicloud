#############      builder                          #############
FROM golang:1.22.5 AS builder

WORKDIR /go/src/github.com/gardener/machine-controller-manager-provider-alicloud
COPY . .

RUN .ci/build

#############      machine-controller               #############
FROM gcr.io/distroless/static-debian11:nonroot AS machine-controller
WORKDIR /

COPY --from=builder /go/src/github.com/gardener/machine-controller-manager-provider-alicloud/bin/rel/machine-controller /machine-controller
ENTRYPOINT ["/machine-controller"]
