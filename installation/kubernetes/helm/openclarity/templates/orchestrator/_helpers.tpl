{{/*
Base name for the orchestrator components
*/}}
{{- define "openclarity.orchestrator.name" -}}
{{- printf "%s-orchestrator" (include  "openclarity.names.fullname" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Kubernetes standard labels
*/}}
{{- define "openclarity.orchestrator.labels.standard" -}}
{{ include "openclarity.labels.standard" . }}
app.kubernetes.io/component: orchestrator
{{- end -}}

{{/*
Labels to use on deploy.spec.selector.matchLabels and svc.spec.selector
*/}}
{{- define "openclarity.orchestrator.labels.matchLabels" -}}
{{ include "openclarity.labels.matchLabels" . }}
app.kubernetes.io/component: orchestrator
{{- end -}}

{{/*
Create the name of the service account to use
*/}}
{{- define "openclarity.orchestrator.serviceAccountName" -}}
{{- if .Values.orchestrator.serviceAccount.create -}}
    {{ default (include "openclarity.orchestrator.name" .) .Values.orchestrator.serviceAccount.name }}
{{- else -}}
    {{ default "default" .Values.orchestrator.serviceAccount.name }}
{{- end -}}
{{- end -}}
