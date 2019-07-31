FROM golang:1.12 as builder
WORKDIR /go/src/github.com/FXinnovation/poweradmin_exporter
COPY . .
RUN make build

FROM alpine:3.9

ARG BUILD_DATE
ARG VCS_REF
ARG VERSION

ENV CA_CERTIFICATES_VERSION=20190108-r0 \
    POWERADMIN_EXPORTER_VERSION=4411b47ff5c61208b1cbd3e8a1e2d097aabdafa7

ADD ./resources /resources

COPY --from=builder /go/src/github.com/FXinnovation/poweradmin_exporter/poweradmin_exporter /resources/poweradmin_exporter

VOLUME /opt/poweradmin_exporter/conf

RUN /resources/build.sh && rm -rf /resources

EXPOSE 9575

USER exporter

VOLUME /opt/poweradmin_exporter/conf

ENTRYPOINT ["/opt/poweradmin_exporter/poweradmin_exporter", "--config.dir=/opt/poweradmin_exporter/conf/"]


LABEL "maintainer"="cloudsquad@fxinnovation.com" \
      "org.label-schema.name"="poweradmin_exporter" \
      "org.label-schema.base-image.name"="docker.io/library/alpine" \
      "org.label-schema.base-image.version"="3.9" \
      "org.label-schema.applications.poweradmin_exporter.version"=${POWERADMIN_EXPORTER_VERSION} \
      "org.label-schema.applications.ca-certificate.version"=${CA_CERTIFICATES_VERSION} \
      "org.label-schema.description"="poweradmin exporter in a container" \
      "org.label-schema.url"="https://github.com/FXinnovation/poweradmin_exporter" \
      "org.label-schema.vcs-url"="https://github.com/FXinnovation/poweradmin_exporter" \
      "org.label-schema.vendor"="FXinnovation" \
      "org.label-schema.schema-version"="0.1.3" \
      "org.label-schema.vcs-ref"=$VCS_REF \
      "org.label-schema.version"=$VERSION \
      "org.label-schema.build-date"=$BUILD_DATE