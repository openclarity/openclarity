{{/*
Base name for the orchestrator components
*/}}
{{- define "vmclarity.orchestrator.name" -}}
{{- printf "%s-orchestrator" (include  "vmclarity.names.fullname" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Kubernetes standard labels
*/}}
{{- define "vmclarity.orchestrator.labels.standard" -}}
{{ include "vmclarity.labels.standard" . }}
app.kubernetes.io/component: orchestrator
{{- end -}}

{{/*
Labels to use on deploy.spec.selector.matchLabels and svc.spec.selector
*/}}
{{- define "vmclarity.orchestrator.labels.matchLabels" -}}
{{ include "vmclarity.labels.matchLabels" . }}
app.kubernetes.io/component: orchestrator
{{- end -}}

{{/*
Create the name of the service account to use
*/}}
{{- define "vmclarity.orchestrator.serviceAccountName" -}}
{{- if .Values.orchestrator.serviceAccount.create -}}
    {{ default (include "vmclarity.orchestrator.name" .) .Values.orchestrator.serviceAccount.name }}
{{- else -}}
    {{ default "default" .Values.orchestrator.serviceAccount.name }}
{{- end -}}
{{- end -}}
