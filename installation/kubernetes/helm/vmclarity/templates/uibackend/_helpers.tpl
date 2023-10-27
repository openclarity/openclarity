{{/*
Base name for the uibackend components
*/}}
{{- define "vmclarity.uibackend.name" -}}
{{- printf "%s-uibackend" (include  "vmclarity.names.fullname" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Kubernetes standard labels
*/}}
{{- define "vmclarity.uibackend.labels.standard" -}}
{{ include "vmclarity.labels.standard" . }}
app.kubernetes.io/component: uibackend
{{- end -}}

{{/*
Labels to use on deploy.spec.selector.matchLabels and svc.spec.selector
*/}}
{{- define "vmclarity.uibackend.labels.matchLabels" -}}
{{ include "vmclarity.labels.matchLabels" . }}
app.kubernetes.io/component: uibackend
{{- end -}}

{{/*
Create the name of the service account to use
*/}}
{{- define "vmclarity.uibackend.serviceAccountName" -}}
{{- if .Values.uibackend.serviceAccount.create -}}
    {{ default (include "vmclarity.uibackend.name" .) .Values.uibackend.serviceAccount.name }}
{{- else -}}
    {{ default "default" .Values.uibackend.serviceAccount.name }}
{{- end -}}
{{- end -}}
