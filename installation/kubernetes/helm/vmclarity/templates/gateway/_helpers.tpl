{{/*
Base name for the gateway components
*/}}
{{- define "vmclarity.gateway.name" -}}
{{- printf "%s-gateway" (include  "vmclarity.names.fullname" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Kubernetes standard labels
*/}}
{{- define "vmclarity.gateway.labels.standard" -}}
{{ include "vmclarity.labels.standard" . }}
app.kubernetes.io/component: gateway
{{- end -}}

{{/*
Labels to use on deploy.spec.selector.matchLabels and svc.spec.selector
*/}}
{{- define "vmclarity.gateway.labels.matchLabels" -}}
{{ include "vmclarity.labels.matchLabels" . }}
app.kubernetes.io/component: gateway
{{- end -}}

{{/*
Create the name of the service account to use
*/}}
{{- define "vmclarity.gateway.serviceAccountName" -}}
{{- if .Values.gateway.serviceAccount.create -}}
    {{ default (include "vmclarity.gateway.name" .) .Values.gateway.serviceAccount.name }}
{{- else -}}
    {{ default "default" .Values.gateway.serviceAccount.name }}
{{- end -}}
{{- end -}}
