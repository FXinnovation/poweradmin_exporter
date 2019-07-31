#!/bin/sh
set -ex

###
# Install poweradmin exporter
###
mkdir -p /opt/poweradmin_exporter/conf
mv /resources/poweradmin_exporter /opt/poweradmin_exporter/

adduser -S -h /opt/poweradmin_exporter exporter

chown -R exporter /opt/poweradmin_exporter

###
# Install dependencies
###
apk add --no-cache \
  ca-certificates=${CA_CERTIFICATES_VERSION}

###
# CIS Hardening
###
touch /etc/login.defs
chmod 0444 /etc/login.defs
chmod 0600 /etc/shadow