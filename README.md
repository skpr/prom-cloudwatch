Prometheus: CloudWatch
======================

Prometheus remote writer for CloudWatch.

![Overview](overview.png)

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