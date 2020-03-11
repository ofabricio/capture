# BUILDER
FROM golang AS builder

ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR $GOPATH/src/

COPY *.go ./

RUN go build \
    -a -tags netgo -ldflags '-w -extldflags "-static"' \
    -o /go/bin/capture \
    *.go


# MAIN
FROM alpine

EXPOSE 9000 9001

COPY --from=builder /go/bin/capture /opt/capture

USER 10001

ENTRYPOINT ["/opt/capture"]

