---
title: Distributed Tracing
weight: 5
---

This page describes how to enable OpenTelemetry distributed tracing for Pipelines-as-Code. When enabled, PaC emits trace spans for webhook event processing and PipelineRun lifecycle timing.

## Enabling tracing

The ConfigMap `pipelines-as-code-config-observability` controls tracing configuration. See [config/305-config-observability.yaml](https://github.com/tektoncd/pipelines-as-code/blob/main/config/305-config-observability.yaml) for the full example.

It contains the following tracing fields:

* `tracing-protocol`: Export protocol. Supported values: `grpc`, `http/protobuf`, `none`. Default is `none` (tracing disabled).
* `tracing-endpoint`: OTLP collector endpoint. Required when protocol is not `none`. The `OTEL_EXPORTER_OTLP_ENDPOINT` environment variable takes precedence if set.
* `tracing-sampling-rate`: Fraction of traces to sample. `0.0` = none, `1.0` = all. Default is `0`.

### Example

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: pipelines-as-code-config-observability
  namespace: pipelines-as-code
data:
  tracing-protocol: grpc
  tracing-endpoint: "http://otel-collector.observability.svc.cluster.local:4317"
  tracing-sampling-rate: "1.0"
```

Changes to the ConfigMap are picked up automatically without restarting the controller. Set `tracing-protocol` to `none` or remove the tracing keys to disable tracing.

## Emitted spans

The controller emits a `PipelinesAsCode:ProcessEvent` span covering the full lifecycle of each webhook event, from receipt through PipelineRun creation. The watcher emits `waitDuration` and `executeDuration` spans for completed PipelineRuns, using the PipelineRun's actual timestamps for accurate wall-clock timing.

## Trace context propagation

When Pipelines-as-Code creates a PipelineRun, it sets the `tekton.dev/pipelinerunSpanContext` annotation with a JSON-encoded OTel TextMapCarrier containing the W3C `traceparent`. PaC tracing works independently — you get PaC spans regardless of whether Tekton Pipelines has tracing enabled.

If Tekton Pipelines is also configured with tracing pointing at the same collector, its reconciler spans appear as children of the PaC span, providing a single end-to-end trace from webhook receipt through task execution. See the [Tekton Pipelines tracing documentation](https://github.com/tektoncd/pipeline/blob/main/docs/developers/tracing.md) for Tekton's independent tracing setup.
