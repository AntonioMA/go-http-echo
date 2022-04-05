# httpEcho

## Description
This module implements a simple server (the simplest possible I think) that just
returns an HTML file with the information it has received (headers, body, request)
for any verb and path.

## Running
To execute the server just compile and run `go-http-echo`. You can also run with a custom template by
executing
```shell
httpEcho -t path/to/template
```

You can create a Docker image with
```shell
make docker-img
```
And the latest version is published as `antonioma/http-echo:latest`

# Template
The template used to format the answer is a normal Go template, extended with the functions on the
[Sprig Library](https://masterminds.github.io/sprig/).

## Changelog
* 0.0.0 Initial version, just the README
* 0.0.1 Initial full version, inline template
* 0.0.2 Add a base-64 encoding of the body
* 0.0.3 Cut the read slice to the actual size read
* 0.0.4 Add support for websockets
* 0.0.5 Unbreak HTML Processing
