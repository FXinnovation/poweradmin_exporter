FROM golang:1.12 as builder
WORKDIR /go/src/github.com/FXinnovation/poweradmin_exporter
COPY . .
RUN make build

FROM quay.io/prometheus/busybox:glibc AS app
LABEL maintainer="FXinnovation CloudToolDevelopment <CloudToolDevelopment@fxinnovation.com>"
COPY --from=builder /go/src/github.com/FXinnovation/poweradmin_exporter/poweradmin_exporter /bin/poweradmin_exporter
EXPOSE      9575
WORKDIR /
ENTRYPOINT  [ "/bin/poweradmin_exporter" ]
