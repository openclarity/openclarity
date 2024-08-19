{{/*
Base name for the gateway components
*/}}
{{- define "openclarity.gateway.name" -}}
{{- printf "%s-gateway" (include  "openclarity.names.fullname" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Kubernetes standard labels
*/}}
{{- define "openclarity.gateway.labels.standard" -}}
{{ include "openclarity.labels.standard" . }}
app.kubernetes.io/component: gateway
{{- end -}}

{{/*
Labels to use on deploy.spec.selector.matchLabels and svc.spec.selector
*/}}
{{- define "openclarity.gateway.labels.matchLabels" -}}
{{ include "openclarity.labels.matchLabels" . }}
app.kubernetes.io/component: gateway
{{- end -}}

{{/*
Create the name of the service account to use
*/}}
{{- define "openclarity.gateway.serviceAccountName" -}}
{{- if .Values.gateway.serviceAccount.create -}}
    {{ default (include "openclarity.gateway.name" .) .Values.gateway.serviceAccount.name }}
{{- else -}}
    {{ default "default" .Values.gateway.serviceAccount.name }}
{{- end -}}
{{- end -}}
