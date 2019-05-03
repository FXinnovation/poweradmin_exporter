# poweradmin_exporter
[![Build Status](https://travis-ci.org/FXinnovation/poweradmin_exporter.svg?branch=master)](https://travis-ci.org/FXinnovation/poweradmin_exporter)

**WARNING - Work In Progress - this exporter is not ready to be used** 

PowerAdmin exporter for Prometheus.

Export metrics from PA - PowerAdmin (https://www.poweradmin.com/) monitoring solution.

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
 	
	 


## Deployment

The exporter listens on port 9575 by default, which can be changed if you need.
[This port is the default port for this exporter.](https://github.com/prometheus/prometheus/wiki/Default-port-allocations)

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

## Contributing

Refer to [CONTRIBUTING.md](https://github.com/FXinnovation/poweradmin_exporter/blob/master/CONTRIBUTING.md).

## License

Apache License 2.0, see [LICENSE](https://github.com/FXinnovation/poweradmin_exporter/blob/master/LICENSE).
