FROM alpine:3.24@sha256:f59abe656b692d901e184be3445e084786602108599d2a51d229c0f5a48c71bf
ARG TARGETPLATFORM
WORKDIR /dashboards
COPY $TARGETPLATFORM/dashboard-linter /usr/local/bin/dashboard-linter
COPY LICENSE /usr/share/licenses/dashboard-linter/LICENSE
USER 65534:65534
ENTRYPOINT ["/usr/local/bin/dashboard-linter"]
