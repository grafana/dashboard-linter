# Grafana Dashboard Linter

This tool is a command-line application to lint Grafana dashboards for common mistakes, and suggest best practices.

## Install

### Prebuilt binaries (recommended)

Download a release archive from the [releases page](https://github.com/grafana/dashboard-linter/releases) and extract the `dashboard-linter` binary onto your `PATH`. Example for Linux on amd64:

```
TAG=v0.4.0
curl -sSfL "https://github.com/grafana/dashboard-linter/releases/download/${TAG}/dashboard-linter_${TAG#v}_linux_amd64.tar.gz" \
  | tar -xz -C /usr/local/bin dashboard-linter
dashboard-linter lint dashboard.json
```

The archive name uses the version without the leading `v`. Each release also publishes a `checksums.txt` you can verify against.

### GitHub Action

Pin the action to a release tag. It downloads that release's binary for the runner's OS and architecture, verifies `checksums.txt`, and adds `dashboard-linter` to `PATH`.

```yaml
- uses: grafana/dashboard-linter@vX.Y.Z
  with:
    files: dashboards
    args: --strict
```

Replace `vX.Y.Z` with the full release tag you want to run. The action is on tags created after this file is merged, and that same tag is the binary that gets installed. The calling workflow only needs `permissions: contents: read`.

`files` is newline-separated dashboard files, directories, or globs such as `dashboards/*.json`. A directory is searched recursively for `*.json` (`node_modules` is skipped), and each file is linted with its own process. With no `--config`, each dashboard still uses a `.lint` file in its own directory, as the binary does. `args` is split on whitespace, so a path with spaces goes in `files` rather than in `args`.

```yaml
permissions:
  contents: read

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@<pinned sha>
      - uses: grafana/dashboard-linter@vX.Y.Z
        with:
          files: |
            dashboards/overview.json
            dashboards/extra.json
          args: --config .lint --strict
```

`--fix` writes files as the job user. `--stdin` is not combined with `files`; omit `files` and run the binary, which is also how to print rules or pass a flag value that contains spaces:

```yaml
- uses: grafana/dashboard-linter@vX.Y.Z
- run: dashboard-linter lint --stdin --strict < dashboards/overview.json
```

Set `version` when the action ref is a branch or commit, or when the binary should be a different published release than the action tag. The `version` output is the installed tag. The `bin` output is the binary path.

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

The prebuilt binaries, GitHub Action, and Docker image above are the supported paths for CI.

This tool is a work in progress and it's still very early days. The current capabilities are focused exclusively on dashboards that use a Prometheus data source.

See [the docs](docs/index.md) for more detail.
