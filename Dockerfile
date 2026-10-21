FROM alpine:3.24@sha256:294b683cb724975bec92580e1e685676bd4b50bda910ddb8c51d4cabeaec77e6
ARG TARGETPLATFORM
WORKDIR /dashboards
COPY $TARGETPLATFORM/dashboard-linter /usr/local/bin/dashboard-linter
COPY LICENSE /usr/share/licenses/dashboard-linter/LICENSE
USER 65534:65534
ENTRYPOINT ["/usr/local/bin/dashboard-linter"]
