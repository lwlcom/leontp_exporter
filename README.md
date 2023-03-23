# leontp_exporter

Prometheus exporter for LeoNTP Devices

# Usage

```
./leontp_exporter
```

# Options

Name     | Default | Description
---------|-------------|----
--version || Print version information
--listen-address | :9330 | Address on which to expose metrics.
--path | /metrics | Path under which to expose metrics.
--config-file |config.yml | File containing"

# Configuration

The config.yml files contains the sensor id and name to use
```
nodes:
  - ntp.example.com
```

## Install
```bash
go get -u github.com/lwlcom/leontp_exporter
```

## Usage
```bash
./leontp_exporter -config-file config.yml
```

## License
(c) Martin Poppen, 2023. Licensed under [MIT](LICENSE) license.

## Prometheus
see https://prometheus.io/
