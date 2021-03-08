[![Maintained by Gruntwork.io](https://img.shields.io/badge/maintained%20by-gruntwork.io-%235849a6.svg)](https://gruntwork.io/?ref=repo_health-checker)

# health-checker

A simple HTTP server that will return `200 OK` if the configured checks are all successful.  If any of the checks fail,
it will return `HTTP 504 Gateway Not Found`.

## Motivation

We were setting up an AWS [Auto Scaling Group](http://docs.aws.amazon.com/autoscaling/latest/userguide/AutoScalingGroup.html)
(ASG) fronted by a [Load Balancer](https://aws.amazon.com/documentation/elastic-load-balancing/) that used a
[Health Check](http://docs.aws.amazon.com/elasticloadbalancing/latest/network/target-group-health-checks.html#) to
determine if the server is healthy. Each server in the ASG runs two services, which means that a server is "healthy" if
the TCP Listeners of both services are successfully accepting connections. But the Load Balancer Health Check is limited to
a single TCP port, or an HTTP(S) endpoint. As a result, our use case just isn't supported natively by AWS.

We wrote health-checker so that we could run a daemon on the server that reports the true health of the server by
attempting to open a TCP connection to more than one port when it receives an inbound HTTP request on the given listener.

Using the `--script` -option, the `health-checker` can be extended to check many other targets. One concrete example is monitoring
`ZooKeeper` node status during rolling deployment. Just polling the `ZooKeeper`'s TCP client port doesn't necessarily guarantee
that the node has (re-)joined the cluster. Using the `health-check` with a custom script target, we can
[monitor ZooKeeper](https://zookeeper.apache.org/doc/r3.4.8/zookeeperAdmin.html#sc_monitoring) using the
[4 letter words](https://zookeeper.apache.org/doc/r3.4.8/zookeeperAdmin.html#sc_zkCommands), ensuring we report health back to the
[Load Balancer](https://aws.amazon.com/documentation/elastic-load-balancing/) correctly.

## How It Works

When health-checker is started, it will listen for inbound HTTP requests for any URL on the IP address and port specified
by `--listener`. When it receives a request, it will attempt to open TCP connections to each of the ports specified by
an instance of `--port` and/or execute the script target specified by `--script`. If all configured checks - all TCP
connections and zero exit status for the script - succeed, it will return `HTTP 200 OK`. If any of the checks fail,
it will return `HTTP 504 Gateway Not Found`.

Configure your AWS Health Check to only pass the Health Check on `HTTP 200 OK`. Now when an HTTP Health Check request
comes in, all desired TCP ports will be checked and the script target executed.

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
| `--script` | Path to script to run - will pass if it completes within configured timeout with a zero exit status. Specify one or more times. | |
| `--script-timeout` | Timeout, in seconds, to wait for the scripts to exit. Applies to all configured script targets. | `5` |
| `--singleflight` | Enables single flight mode, which allows concurrent health check requests to share the results of a single check.  | |
| `--version` | Show the program's version | |

If you execute a shell script, ensure you have a `shebang` line in your script, otherwise the script will fail with an `exec format error`.

#### Example 1

Run a listener on port 6000 that accepts all inbound HTTP connections for any URL. When the request is received,
attempt to open TCP connections to port 5432 and 3306. If both succeed, return `HTTP 200 OK`. If any fails, return `HTTP
504 Gateway Not Found`.

```
health-checker --listener "0.0.0.0:6000" --port 5432 --port 3306
```

#### Example 2

Run a listener on port 6000 that accepts all inbound HTTP connections for any URL. When the request is received,
attempt to open TCP connection to port 5432 and run the script with a 10 second timout. If TCP connection succeeds and script exit code is zero, return `HTTP 200 OK`. If TCP connection fails or non-zero exit code for the script, return `HTTP
504 Gateway Not Found`.

```
health-checker --listener "0.0.0.0:6000" --port 5432 --script /path/to/script.sh --script-timeout 10
```

#### Example 3

Run a listener on port 6000 that accepts all inbound HTTP connections for any URL. When the request is received,
attempt to run the configured scripts. If both return exit code zero, return `HTTP 200 OK`. If either returns non-zero exit code, return `HTTP
504 Gateway Not Found`.

```
health-checker --listener "0.0.0.0:6000" --script "/usr/local/bin/exhibitor-health-check.sh --exhibitor-port 8080" --script "/usr/local/bin/zk-health-check.sh --zk-port 2191"
```
