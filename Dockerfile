FROM alpine:3.24@sha256:be2656061b536c1883c4034790e21d0a09c5aaaf51968b89368a3c0371ea2c9a
ARG TARGETPLATFORM
WORKDIR /dashboards
COPY $TARGETPLATFORM/dashboard-linter /usr/local/bin/dashboard-linter
COPY LICENSE /usr/share/licenses/dashboard-linter/LICENSE
USER 65534:65534
ENTRYPOINT ["/usr/local/bin/dashboard-linter"]
