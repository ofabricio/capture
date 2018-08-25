FROM golang:alpine

COPY . /src
WORKDIR /src

RUN apk --no-cache add git && \
    go get -d ./... && \
    apk del git

# linux | windows | darwin
ENV OS linux
CMD GOOS=$OS GOARCH=386 go build -ldflags="-s -w" -o capture . && \
    chmod -R 777 capture && \
    go version
