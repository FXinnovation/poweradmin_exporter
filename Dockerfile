FROM        quay.io/prometheus/busybox:glibc
LABEL maintainer="FXinnovation CloudToolDevelopment <CloudToolDevelopment@fxinnovation.com>"

COPY poweradmin_exporter /bin/poweradmin_exporter

EXPOSE      9100
ENTRYPOINT  [ "/bin/poweradmin_exporter" ]
