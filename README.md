# Grafana Dashboard Linter

This tool is a command-line application to lint Grafana dashboards for common mistakes, and suggest best practices.

## Install

### Prebuilt binaries (recommended)

Download a release archive from the [releases page](https://github.com/grafana/dashboard-linter/releases) and extract the `dashboard-linter` binary onto your `PATH`. Example for Linux on amd64:

```
VERSION=v0.1.0
curl -sSfL "https://github.com/grafana/dashboard-linter/releases/download/${VERSION}/dashboard-linter_${VERSION}_linux_amd64.tar.gz" \
  | tar -xz -C /usr/local/bin dashboard-linter
dashboard-linter lint dashboard.json
```

Each release also publishes a `checksums.txt` you can verify against.

### From source

```
$ go install github.com/grafana/dashboard-linter@latest
$ dashboard-linter lint dashboard.json
```

Note: `go install ...@<version>` (including `@latest`) currently fails because `go.mod` contains a `replace` directive — Go refuses to install a module with replaces. Build locally with `go build` from a checkout if you need to install from source. The prebuilt binaries above are the supported path for CI.

This tool is a work in progress and it's still very early days. The current capabilities are focused exclusively on dashboards that use a Prometheus data source.

See [the docs](docs/index.md) for more detail.
