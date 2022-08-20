# MSI Afterburner Exporter
convert `MSI Afterburner Remote Server` xml data to Prometheus metrics

usage:
```shell script
go install github.com/wolf-joe/tools/cmd/afterburner_exporter
afterburner_exporter -listen "0.0.0.0:8090" -target "127.0.0.1:82" -password "17cc95b4017d496f82"
```