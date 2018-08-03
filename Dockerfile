FROM golang:1.10 AS builder

RUN go get -u github.com/golang/dep/cmd/dep

RUN mkdir -p /go/src/github.com/helm-helper
WORKDIR /go/src/github.com/helm-helper

ADD . /go/src/github.com/helm-helper

RUN dep ensure --vendor-only
RUN CGO_ENABLED=0 GOOS=linux make build

FROM alpine:latest
COPY --from=builder /go/src/github.com/helm-helper/helm-helper .
CMD ["./helm-helper"]
