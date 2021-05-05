# datadog-monitor-prometheus-exporter
Prometheus Exporter for Datadog monitor.

## Metrics

```
$ curl -s localhost:8080/metrics | grep datadog_monitor_prometheus_exporter
```

## Datadog Autodiscovery

If you use Datadog, you can use [Kubernetes Integration Autodiscovery](https://docs.datadoghq.com/agent/kubernetes/integrations/?tab=kubernetes) feature.


