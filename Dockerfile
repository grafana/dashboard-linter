FROM alpine:3.24
ARG TARGETPLATFORM
WORKDIR /dashboards
COPY $TARGETPLATFORM/dashboard-linter /usr/local/bin/dashboard-linter
COPY LICENSE /usr/share/licenses/dashboard-linter/LICENSE
USER 65534:65534
ENTRYPOINT ["/usr/local/bin/dashboard-linter"]
