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

### Docker image

Each release publishes a multi-arch (linux/amd64 and linux/arm64) image to [GitHub Container Registry](https://github.com/grafana/dashboard-linter/pkgs/container/dashboard-linter). The image's working directory is `/dashboards`, so mounting your dashboards directory there lets you lint by relative path:

```
docker run --rm -v "$PWD:/dashboards" ghcr.io/grafana/dashboard-linter:latest lint dashboard.json
```

Pin a version and use `--strict` to fail your pipeline on lint warnings:

```
docker run --rm -v "$PWD:/dashboards" ghcr.io/grafana/dashboard-linter:v0.3.0 lint dashboard.json --strict
```

To lint from stdin:

```
cat dashboard.json | docker run --rm -i ghcr.io/grafana/dashboard-linter:latest lint --stdin
```

The container runs as an unprivileged user (UID 65534) and cannot write to the mounted directory. If you use `--fix`, run as your own user so the autofixed files are not owned by root:

```
docker run --rm --user "$(id -u):$(id -g)" -v "$PWD:/dashboards" ghcr.io/grafana/dashboard-linter:latest lint dashboard.json --fix
```

The image defines an entrypoint, so CI systems that prepend their own commands need to clear it. GitLab CI:

```yaml
lint-dashboards:
  image:
    name: ghcr.io/grafana/dashboard-linter:v0.3.0
    entrypoint: [""]
  script:
    - dashboard-linter lint dashboard.json --strict
```

### From source

`go install github.com/grafana/dashboard-linter@<version>` does not currently work because `go.mod` contains a `replace` directive — Go refuses to install a module with replaces. Build from a checkout instead:

```
$ git clone https://github.com/grafana/dashboard-linter.git
$ cd dashboard-linter
$ go build -o dashboard-linter .
$ ./dashboard-linter lint dashboard.json
```

The prebuilt binaries and Docker image above are the supported paths for CI.

This tool is a work in progress and it's still very early days. The current capabilities are focused exclusively on dashboards that use a Prometheus data source.

See [the docs](docs/index.md) for more detail.
