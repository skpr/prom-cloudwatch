Prometheus: CloudWatch
======================

Prometheus remote writer for CloudWatch.

![Overview](docs/overview.png)

## Usage

**Run prometheus-cloudwatch**

```bash
$ ./prometheus-cloudwatch 
INFO[0000] Starting writer server: :8080                 source="main.go:131"
INFO[0000] Starting metrics servere: :9000               source="main.go:152"
```

**Configure Prometheus**

```yaml
remote_write:
  - url: http://storage:8080/write
```
