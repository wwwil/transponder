# Transponder

Transponder is a continuously running multi-protocol network connectivity testing utility for Kubernetes and Istio.

[![Docker Repository on Quay](https://quay.io/repository/wwwil/transponder/status "Docker Repository on Quay")](https://quay.io/repository/wwwil/transponder)

:construction:
:warning:
**This project is currently just a proof of concept.**
:warning:
:construction:

## Background

Transponder is a simple application written in Go. It has two parts; a _server_ and a _scanner_. The server exposes
endpoints for HTTP, HTTPS, and gRPC. The scanner will then connect to these ports using the corresponding protocol to
verify that a request can be made.

Multiple servers can be deployed in different places and the scanner will continuously cycle through the servers and
ports and repeatedly attempt to make connections. By watching its log output it's possible to see when connections fail.

This can then be run while setting up networking to verify connectivity and quickly see when a breaking change occurs.

## Use Cases

Possible uses for Transponder include:
- Checking external connectivity to a cluster to verify a load balancer is working.
- Checking cross cluster connectivity in a mutli cluster Istio setup.
- Checking that undesired connectivity is not possible inside a cluster.

## Local Usage

Transponder binaries can be found on the [releases page](https://github.com/wwwil/transponder/releases). Download and
`unzip` the latest correct binary for your OS and architecture.

The server can be run like so:

```bash
./transponder server
```

By default the server uses port `8080` for HTTP, `8081` for gRPC, and `8443` for HTTPS. These can be changed using
arguments, for example `--http-port=1234`.

The scanner is then run like this:

```bash
./transponder scanner
```

It uses a configuration file to specify a list of servers and ports to connect to. By default it looks for a file called
`scanner.yaml` in the current directory. This can be changed using the argument
`--config-file=~/my-scanner-config.yaml`. An example configuration file can be found in [`scanner.yaml`](scanner.yaml).

## Container Usage

Transponder container images can be found on the [Quay repository](https://quay.io/repository/wwwil/transponder):

```bash
docker pull quay.io/wwwil/transponder
```

The server can be deployed to a Kubernetes cluster using the
[server example manifest](deployments/transponder-server.yaml):

```bash
kubectl apply -f ./deployments/transponder-server.yaml
```

The scanner can either be run locally against a remote server, or also be deployed into the cluster using the
[scanner example manifest](deployments/transponder-scanner.yaml):

```bash
kubectl apply -f ./deployments/transponder-scanner.yaml
```

## Demo

There is a [demo](demo) to shown how Transponder can help when configuring Kubernetes and Istio. 

## Building

All building is handled by the [`make.sh`](make.sh) bash script which runs the function named by the argument passed to it.

To build a Transponder binary run:

```bash
./make.sh build
```

To build a Docker image of Transponder:

```bash
./make.sh docker-build
```

## Roadmap

- Build and push official image so users don't have to build it themselves.
- Add support for other protocols (TCP, TLS, etc.) using the list of
[Istio protocols](https://istio.io/latest/docs/ops/configuration/traffic-management/protocol-selection/) as a guide.
- Add a central observation component to aggregate logs from multiple scanners.

## Motivation

While working on a Jetstack project to set up Istio on a customer's clusters we often found it hard to establish with
confidence that services could communicate properly in cluster, across clusters, and externally. When there were issues
it was sometimes hard to establish whether they were a general configuration problem or something more specific to how
the customer’s services worked.

To help with this we deployed various HTTP demo applications which could be `curl`ed from other `Pods`, or from outside
the clusters. We also used the
[gRPC greeter app](https://github.com/GoogleCloudPlatform/istio-samples/tree/master/sample-apps/grpc-greeter-go)
provided by Istio to verify that gRPC was working without having to get too involved with the customer's existing
services.

Transponder aims to streamline this kind of testing and make it easy to check connectivity of different protocols
between two points.
