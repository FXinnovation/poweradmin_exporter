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

## Deployment

The exporter listens on port 9575 by default, which can be changed if you need.
[This port is the default port for this exporter.](https://github.com/prometheus/prometheus/wiki/Default-port-allocations)

## Contributing

Refer to [CONTRIBUTING.md](https://github.com/FXinnovation/poweradmin_exporter/blob/master/CONTRIBUTING.md).

## License

Apache License 2.0, see [LICENSE](https://github.com/FXinnovation/poweradmin_exporter/blob/master/LICENSE).
