# health-checker

A simple HTTP server that will return `200 OK` if the configured checks are all successful.

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

When health-checker is started, it will parse a YAML file specified with the `--config` flag (example config
in [examples/config.yml.simple]()) and listen for inbound HTTP requests for any URL on the IP address and port specified
by `listener` directive. When it receives a request, it will attempt to run all checks specified in the config
and return `HTTP 200 OK` if all checks pass. If any of the checks fail, it will return `HTTP 504 GATEWAY TIMEOUT`.

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
| `--config` | A YAML config file containing options and checks | |
| `--log-level` | Set the log level to LEVEL. Must be one of: `panic`, `fatal`, `error,` `warning`, `info`, or `debug` | `info` 
| `--help` | Show the help screen | | 
| `--version` | Show the program's version | | 

#### Config File Options

TODO: add more info on the config options

#### Examples

Parse configuration from `health-checker.yml` and run a listener  that accepts all inbound HTTP connections for any URL. When
the request is received, attempt to run all checks specified in `health-checker.yml`. If all checks succeed, return `HTTP 200 OK`. 
If any fail, return `HTTP 504 GATEWAY TIMEOUT`.

```
health-checker --config health-checker.yml
```

See [examples/]() for configuration examples.
