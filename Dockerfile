FROM golang:latest
COPY . /src
WORKDIR /src
RUN go get -d ./...
# linux | windows | darwin
ENV OS linux
CMD GOOS=$OS GOARCH=386 go build -ldflags="-s -w" -o capture . && \
    chmod -R 777 capture
