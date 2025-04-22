# link-exporter
Prometheus exporter to gather latency metrics between two physical links.

# Usage

Exporter initiates client/server connection between two physical links using link-local IPv6 addresses.
See `config_example.json` for available options.

To compile exporter run:
```
go build -o link-exporter cmd/link-exporter/main.go
```
or
```
CGO_ENABLED=0 go build -o link-exporter cmd/link-exporter/main.go
```
To compile cross-distribution version.

Execute exporter by running:

```
./link-exporter -config_file ./config.json
```
