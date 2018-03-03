# health-checker

A simple HTTP server that will return `200 OK` if the configured checks are all successful.

## Quick start

Put the following in `health-checks.yml`:

```yaml
tcp:
  - name: tcpService1
    host: localhost
    port: 5500
http:
  - name: httpService1
    host: 127.0.0.1
    port: 8080
    success_codes: [200, 204, 301, 302]
```

and run `health-checker`. Now, requests to `0.0.0.0:5000` will return a 200 OK if all the checks
specified in `health-checks.yml` pass.

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

When `health-checker` is started, it will parse a YAML file specified with the `--config` option (see [examples folder](examples/)) 
and listen for inbound HTTP requests for any URL on the IP address and port specified by `--listener`. When it receives a request, 
it will evaluate all checks and return `HTTP 200 OK` if all checks pass. If any of the checks fail, it will return `HTTP 504 GATEWAY TIMEOUT`.

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
| `--listener` |  The IP address and port on which inbound HTTP connections will be accepted. | `0.0.0.0:5000`
| `--config` | A YAML file containing checks which will be evaluated | `health-checks.yml`
| `--log-level` | Set the log level to LEVEL. Must be one of: `panic`, `fatal`, `error,` `warning`, `info`, or `debug` | `info` 
| `--help` | Show the help screen | | 
| `--version` | Show the program's version | | 

```
health-checker --listener "0.0.0.0:6000" --config "my-checks.yml" --log-level "warning"
```

##### Examples

See [examples folder](examples/) for more complete `health-checks.yml` examples.
