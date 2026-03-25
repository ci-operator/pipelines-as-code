package reconciler

import (
	"context"
	"encoding/json"

	"github.com/openshift-pipelines/pipelines-as-code/pkg/apis/pipelinesascode/keys"
	"github.com/openshift-pipelines/pipelines-as-code/pkg/tracing"
	tektonv1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.opentelemetry.io/otel/trace"
	corev1 "k8s.io/api/core/v1"
	"knative.dev/pkg/apis"
)

const (
	applicationLabel = "appstudio.openshift.io/application"
	componentLabel   = "appstudio.openshift.io/component"
	stageBuild       = "build"
)


// extractSpanContext extracts the trace context from the pipelinerunSpanContext annotation.
func extractSpanContext(pr *tektonv1.PipelineRun) (context.Context, bool) {
	raw, ok := pr.GetAnnotations()[keys.SpanContextAnnotation]
	if !ok || raw == "" {
		return nil, false
	}
	var carrierMap map[string]string
	if err := json.Unmarshal([]byte(raw), &carrierMap); err != nil {
		return nil, false
	}
	carrier := propagation.MapCarrier(carrierMap)
	ctx := otel.GetTextMapPropagator().Extract(context.Background(), carrier)
	sc := trace.SpanContextFromContext(ctx)
	if !sc.IsValid() {
		return nil, false
	}
	return ctx, true
}

// emitTimingSpans emits wait_duration and execute_duration spans for a completed build PipelineRun.
func emitTimingSpans(pr *tektonv1.PipelineRun) {
	parentCtx, ok := extractSpanContext(pr)
	if !ok {
		return
	}

	tracer := otel.Tracer(tracing.TracerName)
	commonAttrs := buildCommonAttributes(pr)

	// Emit waitDuration: creationTimestamp -> status.startTime
	if pr.Status.StartTime != nil {
		_, waitSpan := tracer.Start(parentCtx, "waitDuration",
			trace.WithTimestamp(pr.CreationTimestamp.Time),
			trace.WithAttributes(commonAttrs...),
		)
		waitSpan.End(trace.WithTimestamp(pr.Status.StartTime.Time))
	}

	// Emit executeDuration: status.startTime -> status.completionTime
	if pr.Status.StartTime != nil && pr.Status.CompletionTime != nil {
		execAttrs := append(append([]attribute.KeyValue{}, commonAttrs...), buildExecuteAttributes(pr)...)
		_, execSpan := tracer.Start(parentCtx, "executeDuration",
			trace.WithTimestamp(pr.Status.StartTime.Time),
			trace.WithAttributes(execAttrs...),
		)
		execSpan.End(trace.WithTimestamp(pr.Status.CompletionTime.Time))
	}
}

// buildCommonAttributes returns span attributes common to both timing spans.
func buildCommonAttributes(pr *tektonv1.PipelineRun) []attribute.KeyValue {
	attrs := []attribute.KeyValue{
		semconv.K8SNamespaceName(pr.GetNamespace()),
		tracing.TektonPipelineRunNameKey.String(pr.GetName()),
		tracing.TektonPipelineRunUIDKey.String(string(pr.GetUID())),
		tracing.DeliveryStageKey.String(stageBuild),
		tracing.DeliveryApplicationKey.String(pr.GetLabels()[applicationLabel]),
	}
	if component := pr.GetLabels()[componentLabel]; component != "" {
		attrs = append(attrs, tracing.DeliveryComponentKey.String(component))
	}
	return attrs
}

// buildExecuteAttributes returns span attributes specific to execute_duration.
func buildExecuteAttributes(pr *tektonv1.PipelineRun) []attribute.KeyValue {
	cond := pr.Status.GetCondition(apis.ConditionSucceeded)
	success := false
	reason := ""
	if cond != nil {
		reason = cond.Reason
		success = cond.Status == corev1.ConditionTrue
	}
	return []attribute.KeyValue{
		tracing.DeliverySuccessKey.Bool(success),
		tracing.DeliveryReasonKey.String(reason),
	}
}
