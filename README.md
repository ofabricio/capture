
**Capture** is a reverse proxy that takes an incoming HTTP request and sends it to another server,
proxying the response back to the client, while showing them in a dashboard.

[![Build Status](https://github.com/ofabricio/capture/workflows/build/badge.svg)](https://github.com/ofabricio/capture/actions?query=workflow%3Abuild)
[![Github Release](https://img.shields.io/github/release/ofabricio/capture.svg)](https://github.com/ofabricio/capture/releases)


## Running

```sh
./capture -url=https://example.com/
```

#### Settings

| param        | description |
|--------------|-------------|
| `-url`       | **Required.** Set the url to proxy |
| `-port`      | Set the proxy port. Default: *9000* |
| `-dashboard` | Set the dashboard port. Default: *9001* |


## Using

If you set your base url as `http://example.com/api`, now `http://localhost:9000` points to that
address. Hence, calling `http://localhost:9000/users/1` is like calling `http://example.com/api/users/1`

*Capture* saves all requests and responses so that you can see them in the dashboard.


## Dashboard

To access the dashboard go to `http://localhost:9001/`

##### Preview

![dashboard](https://i.imgur.com/4yOSWFn.png)


## Building

### For manual build

```sh
git clone --depth 1 https://github.com/ofabricio/capture.git
cd capture
go build
```

### For building with docker

```sh
git clone --depth 1 https://github.com/ofabricio/capture.git
cd capture
docker run --rm -v $PWD:/src -w /src -e GOOS=darwin -e GOARCH=amd64 golang:alpine go build
```

Now you have an executable binary in your directory.

**Note:** set `GOOS=darwin` to `linux` or `windows` to create an executable for the corresponding Operating System.

### For running straight from docker

```sh
git clone --depth 1 https://github.com/ofabricio/capture.git
cd capture
docker run --rm -v $PWD:/src -w /src golang:alpine apk add ca-certificates && go run main.go -url=http://example.com
```
