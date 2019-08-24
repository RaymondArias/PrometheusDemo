# Prometheus Metric Demo
A simple web application written in Go with Prometheus metrics.

There are 4 endpoints that do various functions. Each with metrics to record request latency and a counter.

```
/health - returns a 200 "OK" message, used for k8s liveness and readiness probe
/base64?data="<something> - returns the base64 encoding of data
/sayhi?name=<name> - returns "Hello <name>!", and increments name in redis
/count?name=<name> - returns the number of times <name> has been called for sayhi
```

The current deployment works with 
- [Prometheus Operator](https://github.com/coreos/prometheus-operator)
- [Datadog Prometheus Intregation](https://docs.datadoghq.com/getting_started/integrations/prometheus?tab=kubernetes)

## Deployment
`helm install ./helm/prometheus
