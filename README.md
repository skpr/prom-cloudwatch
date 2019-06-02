Prometheus: CloudWatch [![CircleCI](https://circleci.com/gh/skpr/prometheus-cloudwatch.svg?style=svg)](https://circleci.com/gh/skpr/prometheus-cloudwatch)
======================

Prometheus remote writer for CloudWatch.

![Overview](docs/overview.png)

## Usage

**Run prometheus-cloudwatch**

```bash
$ ./prometheus-cloudwatch 
INFO[0000] Starting server: :8080    source="main.go:102"
```

**Configure Prometheus**

```yaml
remote_write:
  - url: http://storage:8080/write
```
