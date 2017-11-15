
**Capture** is a reverse proxy that captures the network traffic and shows it in a dashboard


## Building / Running

    git clone https://github.com/ofabricio/capture.git
    cd capture
    go build
    ./capture -url=https://example.com/api -port=9000 -dashboard=apple -max-captures=16

### Binaries / Executables

For ready-to-use executables (no need to build it yourself) for *Windows* and *Linux*, see [Releases](https://github.com/ofabricio/capture/releases) page

### Configurations

| param           | description |
|-----------------|-------------|
| `-url`          | **Required.** Set the base url you want to capture |
| `-port`         | Set the port you want to capture. Default: *9000* |
| `-dashboard`    | Set the dashboard name. Default: *dashboard* |
| `-max-captures` | Set the max number of captures. Default: *16* |
| `-h`            | Show help |


## Using

If you set your base url as `http://example.com/api`, now `http://localhost:9000` points to that
address. Hence, calling `http://localhost:9000/users/1` is like calling `http://example.com/api/users/1`

*Capture* saves all requests and responses so that you can see them in the dashboard


## Dashboard

To access the dashboard go to `http://localhost:9000/dashboard`

The path `/dashboard/**` is reserved, that means if your api has a path like that it will be ignored
in favor of the dashboard. However, you can change the dashboard's name with `-dashboard`

##### Preview

![dashboard](https://i.imgur.com/13nzb48.png)