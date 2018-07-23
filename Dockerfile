FROM golang:latest
COPY . /src
WORKDIR /src
RUN go get -d ./...
ENV OS linux # linux | windows | darwin
CMD GOOS=$OS GOARCH=386 go build -ldflags="-s -w" -o capture . && \
    chmod -R 777 capture
