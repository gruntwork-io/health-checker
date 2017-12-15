# health-checker

A simple HTTP server that will return `200 OK` if the given TCP ports are all successfully accepting connections.

## Motivation

We were setting up an AWS [Auto Scaling Group](http://docs.aws.amazon.com/autoscaling/latest/userguide/AutoScalingGroup.html)
(ASG) fronted by a [Load Balancer](https://aws.amazon.com/documentation/elastic-load-balancing/) that used a 
[Health Check](http://docs.aws.amazon.com/elasticloadbalancing/latest/network/target-group-health-checks.html#) to
determine if the server is healthy. Each server in the ASG runs two services, which means that a server is "healthy" if
the TCP Listeners of both services are successfully accepting connections. But the Load Balancer Health Check is limited to 
a single TCP port, or an HTTP(S) endpoint. As a result, our use case just isn't supported natively by AWS.

We wrote health-checker so that we could run a daemon on the server that reports the true health of the server by
attempting to open a TCP connection to more than one port when it receives an inbound HTTP request on the given listener.

## How It Works

When health-checker is started, it will listen for inbound HTTP requests for any URL on the IP address and port specified
by `--listener`. When it receives a request, it will attempt to open TCP connections to each of the ports specified by
an instance of `--port`. If all TCP connections succeed, it will return `HTTP 200 OK`. If any TCP connection fails, it
will return `HTTP 504 Gateway Not Found`. 

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
| `--port` | The port number on which a TCP connection will be attempted. Specify one or more times. | | 
| `--listener` |  The IP address and port on which inbound HTTP connections will be accepted. | `0.0.0.0:5000`
| `--log-level` | Set the log level to LEVEL. Must be one of: `panic`, `fatal`, `error,` `warning`, `info`, or `debug` | `info` 
| `--help` | Show the help screen | | 
| `--version` | Show the program's version | | 

#### Example

Run a listener on port 6000 that accepts all inbound HTTP connections for any URL. When the request is received,
attempt to open TCP connections to port 5432 and 3306. If both succeed, return `HTTP 200 OK`. If any fails, return `HTTP
504 Gateway Not Found`.

```
health-checker --listener "0.0.0.0:6000" --port 5432 --port 3306
```

