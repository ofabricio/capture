
**Capture** is a reverse proxy that captures the network traffic and shows it in a dashboard

[![Build Status](https://travis-ci.com/ofabricio/capture.svg?branch=master)](https://travis-ci.com/ofabricio/capture)


## Binaries / Executables

For ready-to-use executables for *Windows*, *Linux* and *Mac*, see [Releases](https://github.com/ofabricio/capture/releases) page


## Running

    ./capture -url=https://example.com/


### Configurations

| param           | description |
|-----------------|-------------|
| `-url`          | **Required.** Set the base url you want to capture |
| `-port`         | Set the proxy port. Default: *9000* |
| `-dashboard`    | Set the dashboard's name. Default: *dashboard* |
| `-max-captures` | Set the max number of captures to show in the dashboard. Default: *16* |


## Using

If you set your base url as `http://example.com/api`, now `http://localhost:9000` points to that
address. Hence, calling `http://localhost:9000/users/1` is like calling `http://example.com/api/users/1`

*Capture* saves all requests and responses so that you can see them in the dashboard


## Dashboard

To access the dashboard go to `http://localhost:9000/dashboard`

The path `/dashboard/**` is reserved, that means if your api has a path like that it will be ignored
in favor of the dashboard. However, you can change the dashboard's name with `-dashboard`


##### Preview

![dashboard](https://i.imgur.com/V2mEUfZ.png)


## Building

Manually:

    git clone https://github.com/ofabricio/capture.git
    cd capture
    go build -o capture .

Via docker:

    git clone https://github.com/ofabricio/capture.git
    cd capture
    docker run --rm -v "${PWD}:/src" -w /src -e GOOS=darwin -e GOARCH=386 golang:latest go build -ldflags="-s -w" -o capture .

Now you have an executable binary in your folder

**Note:** you can change `GOOS=darwin` to `linux` or `windows`
