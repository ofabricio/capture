
**Capture** is a reverse proxy that takes an incoming HTTP request and sends it to another server,
proxying the response back to the client, while showing them in a dashboard.

[![Build Status](https://github.com/ofabricio/capture/workflows/build/badge.svg)](https://github.com/ofabricio/capture/actions?query=workflow%3Abuild)
[![Github Release](https://img.shields.io/github/release/ofabricio/capture.svg)](https://github.com/ofabricio/capture/releases)


## Running

    ./capture -url=https://example.com/


#### Settings

| param        | description |
|--------------|-------------|
| `-url`       | **Required.** Set the url you want to proxy |
| `-port`      | Set the proxy port. Default: *9000* |
| `-dashboard` | Set the dashboard port. Default: *9001* |
| `-captures`  | Set how many captures to show in the dashboard. Default: *16* |


## Using

If you set your base url as `http://example.com/api`, now `http://localhost:9000` points to that
address. Hence, calling `http://localhost:9000/users/1` is like calling `http://example.com/api/users/1`

*Capture* saves all requests and responses so that you can see them in the dashboard.


## Dashboard

To access the dashboard go to `http://localhost:9001/`

##### Preview

![dashboard](https://i.imgur.com/4yOSWFn.png)


## Building

Manually:

    git clone --depth 1 https://github.com/ofabricio/capture.git
    cd capture
    go build

Via docker:

    git clone --depth 1 https://github.com/ofabricio/capture.git
    cd capture
    docker run --rm -v $PWD:/src -w /src -e GOOS=darwin -e GOARCH=amd64 golang:alpine go build

Now you have an executable binary in your directory.

**Note:** change `GOOS=darwin` to `linux` or `windows` to create an executable for your corresponding Operating System. For linux, you also need to change the image tag to `golang:1.16`.

## Plugins

Put [plugin](https://golang.org/pkg/plugin/) files in the current directory.
They are loaded sorted by filename on startup.

Plugins must export the following function:

```go
func Handler(proxy http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        proxy(w, r)
    }
}
```
