# health-checker

A simple HTTP server that will return `200 OK` if the given checks are all successful.

## Motivation

We were setting up an AWS [Auto Scaling Group](http://docs.aws.amazon.com/autoscaling/latest/userguide/AutoScalingGroup.html)
(ASG) fronted by a [Load Balancer](https://aws.amazon.com/documentation/elastic-load-balancing/) that used a 
[Health Check](http://docs.aws.amazon.com/elasticloadbalancing/latest/network/target-group-health-checks.html#) to
determine if the server is healthy. Each server in the ASG runs two services, which means that a server is "healthy" if
the TCP Listeners of both services are successfully accepting connections. But the Load Balancer Health Check is limited to 
a single TCP port, or an HTTP(S) endpoint. As a result, our use case just isn't supported natively by AWS.

We wrote health-checker so that we could run a daemon on the server that reports the true health of the server by
checking more conditions than a just single port or HTTP request while still allowing for a single HTTP request on the given listener.

## How It Works

When health-checker is started, it will listen for inbound HTTP requests for any URL on the IP address and port specified
by `--listener`. When it receives a request, it will attempt to open TCP connections to each of the ports specified by
an instance of `--port`, send a request out to each of the HTTP endpoints specified by `--http`, and run all scripts
specified by `--script`. If all TCP connections succeed, HTTP requests return a 2XX status code, and all specified scripts return
with a zero exit code, it will return `HTTP 200 OK`. If any of the specified checks fail, it will return `HTTP 504 Gateway Not Found`.

Configure your AWS Health Check to only pass the Health Check on `HTTP 200 OK`. Now when an HTTP Health Check request
comes in, all desired TCP ports will be checked.

For stability, we recommend running health-checker under a process supervisor such as [supervisord](http://supervisord.org/)
or [systemd](https://www.freedesktop.org/wiki/Software/systemd/) to automatically restart health-checker in the unlikely
case that it fails.

## Installation

Just download the latest release for your OS on the [Releases Page](https://github.com/gruntwork-io/health-checker/releases).

## Usage

```
health-checker [options]
```

#### Options

| Option | Description | Default 
| ------ | ----------- | -------
| `--port` | The port number on which a TCP connection will be attempted. Can be specified multiple times. | |
| `--http` | The url:port to check for a 2XX status code. Can be specified multiple times. | |
| `--script` | Path to an executable script that should return with a 0 exit status if successful. Can be specified multiple times. | |
| `--listener` |  The IP address and port on which inbound HTTP connections will be accepted. | `0.0.0.0:5000`
| `--log-level` | Set the log level to LEVEL. Must be one of: `panic`, `fatal`, `error,` `warning`, `info`, or `debug` | `info` 
| `--help` | Show the help screen | | 
| `--version` | Show the program's version | | 

#### Examples

Run a listener on port 6000 that accepts all inbound HTTP connections for any URL. When the request is received,
attempt to open TCP connections to port 5432 and 3306. If both succeed, return `HTTP 200 OK`. If any fail, return `HTTP
504 Gateway Not Found`.

```
health-checker --listener "0.0.0.0:6000" --port 5432 --port 3306
```

Run a listener on port 6000 that accepts all inbound HTTP connections for any URL. When the request is received,
attempt to open TCP connections to port 5432 and send an HTTP request to `localhost:80`. If a connection is successfully opened
to port 5432 and the service at `localhost:80` responds with a 2XX status code, return `HTTP 200 OK`. If any fail, return `HTTP
504 Gateway Not Found`.

```
health-checker --listener "0.0.0.0:6000" --port 5432 --http "localhost:80"
```

Run a listener on port 6000 that accepts all inbound HTTP connections for any URL. When the request is received,
attempt to open TCP connections to port 5432, send an HTTP request to `localhost:80`, and run the script at `/usr/local/bin/check_foo.sh`.
If a connection is successfully opened to port 5432, the service at `localhost:80` responds with a 2XX status code, and the script
exits with a zero exit status code, return `HTTP 200 OK`. If any fail, return `HTTP 504 Gateway Not Found`.

```
health-checker --listener "0.0.0.0:6000" --port 5432 --http "localhost:80" --script "/usr/local/bin/check_foo.sh"
```
