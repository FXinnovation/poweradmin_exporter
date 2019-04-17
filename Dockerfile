FROM        quay.io/prometheus/busybox:glibc
LABEL maintainer="FXinnovation CloudToolDevelopment <CloudToolDevelopment@fxinnovation.com>"

COPY exporter-template /bin/exporter-template

EXPOSE      9100
ENTRYPOINT  [ "/bin/exporter-template" ]
