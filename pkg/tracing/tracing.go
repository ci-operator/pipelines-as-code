package tracing

import "go.opentelemetry.io/otel/attribute"

const TracerName = "pipelines-as-code"

// Span attribute keys.
var (
	DeliveryStageKey       = attribute.Key("delivery.stage")
	DeliveryApplicationKey = attribute.Key("delivery.application")
	DeliveryComponentKey   = attribute.Key("delivery.component")
	DeliverySuccessKey     = attribute.Key("delivery.success")
	DeliveryReasonKey      = attribute.Key("delivery.reason")

	TektonPipelineRunNameKey = attribute.Key("tekton.pipelinerun.name")
	TektonPipelineRunUIDKey  = attribute.Key("tekton.pipelinerun.uid")

	VCSEventTypeKey  = attribute.Key("vcs.event_type")
	VCSProviderKey   = attribute.Key("vcs.provider")
	VCSRepositoryKey = attribute.Key("vcs.repository.url")
	VCSRevisionKey   = attribute.Key("vcs.revision")
)
