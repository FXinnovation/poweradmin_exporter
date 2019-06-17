# poweradmin_exporter
[![Build Status](https://travis-ci.org/FXinnovation/poweradmin_exporter.svg?branch=master)](https://travis-ci.org/FXinnovation/poweradmin_exporter)


PowerAdmin exporter for Prometheus.

Export metrics from PA - PowerAdmin (https://www.poweradmin.com/) monitoring solution using the Power Admin API.
It takes the status of the monitors and exposes them as an exporter. The status is exposed and mapped to a table which is dynamic.
### Features
- Uses the PA API to fetch monitor values
- You can map the monitor status to a number in a file
- Transform names of monitors so that it's compliant with prometheus
### Prerequisites

To run this project, you will need a [working Go environment](https://golang.org/doc/install).

### Installing

```bash
go get -u github.com/FXinnovation/poweradmin_exporter
```

## Running the tests

```bash
make test
```

## Usage

```bash
./poweradmin_exporter -h
```

## PowerAdmin integration
### Status mapping for monitors
PowerAdmin API uses the following values for the status of the monitors.

Status|Value in GET_MONITOR_INFO call
------|-----
Alert|2
Alert - Skipping Actions|10
Alert - Green|17
Alert - Red|18
Alert - Suppressing|19
Bad License|14
Can't Run|4
Dependency Not Met|16
Disabled|6
Error|3
Error - Suppressed|21
OK|1
OK - Unacknowledged Alerts - Yellow|20
OK - Unacknowledged Alerts - Red|24
OK - Unacknowledged Alerts - Green|25
Monitor Busy|11
Monitor Maintenance Mode|13
Satellite Disconnected|23
Scheduled|7
Server Disabled|22
Server Maintenance Mode|26
Startup Pause|8
Training|12
Unlicensed|9
 	

The _status_mapping.yml_ file contains the values as float64 returned in the metrics for each string status returned by PowerAdmin. You can also specify a default value.
	 


## Deployment

The exporter listens on port 9575 by default, which can be changed if you need.
[This port is the default port for this exporter.](https://github.com/prometheus/prometheus/wiki/Default-port-allocations)

### Exporter configuration

A config.yml file must exist and contains the following items
```
server: "https://paserver"
api_key: "THE_API_KEY"
skip_tls_verify: false
group: ## a list of group names to monitor
  - name: "Dev"
    servers:
      - "Server"
  - name: "MyWonderfulMachines"
```
The _skip_tls_verify_ option gives you the possibility to skip the certificate checking for self signed certs for example.
## Building
Build the sources with 
```bash
make build
```
**Note**: As this is a go build you can use _GOOS_ and _GOARCH_ environment variables to build for another platform.
### Crossbuilding
The _Makefile_ contains a _crossbuild_ target which builds all the platforms defined in _.promu.yml_ file and puts the files in _.build_ folder. Alternatively you can specify one platform to build with the OSARCH environment variable;
```bash
OSARCH=linux/amd64 make crossbuild
```
## Docker image

You can build a docker image using:
```bash
make docker
```
The resulting image is named `fxinnovation/poweradmin_exporter:{git-branch}`.
It exposes port 9575 and expects the config in /config.yml. To configure it, you can bind-mount a config from your host: 
```
$ docker run -p 9575 -v /path/on/host/config.yml:/config.yml fxinnovation/poweradmin_exporter:master
```
By default, the status_mapping.yml file is included in the docker image but you can override it and use a new one with the same command as the config.yml:
```
$ docker run -p 9575 -v /path/on/host/config.yml:/config.yml -v /path/on/host/status_mapping.yml:/status_mapping.yml fxinnovation/poweradmin_exporter:master
```

## Contributing

Refer to [CONTRIBUTING.md](https://github.com/FXinnovation/poweradmin_exporter/blob/master/CONTRIBUTING.md).

## License

Apache License 2.0, see [LICENSE](https://github.com/FXinnovation/poweradmin_exporter/blob/master/LICENSE).
