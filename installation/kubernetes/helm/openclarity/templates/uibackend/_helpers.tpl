{{/*
Base name for the uibackend components
*/}}
{{- define "openclarity.uibackend.name" -}}
{{- printf "%s-uibackend" (include  "openclarity.names.fullname" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Kubernetes standard labels
*/}}
{{- define "openclarity.uibackend.labels.standard" -}}
{{ include "openclarity.labels.standard" . }}
app.kubernetes.io/component: uibackend
{{- end -}}

{{/*
Labels to use on deploy.spec.selector.matchLabels and svc.spec.selector
*/}}
{{- define "openclarity.uibackend.labels.matchLabels" -}}
{{ include "openclarity.labels.matchLabels" . }}
app.kubernetes.io/component: uibackend
{{- end -}}

{{/*
Create the name of the service account to use
*/}}
{{- define "openclarity.uibackend.serviceAccountName" -}}
{{- if .Values.uibackend.serviceAccount.create -}}
    {{ default (include "openclarity.uibackend.name" .) .Values.uibackend.serviceAccount.name }}
{{- else -}}
    {{ default "default" .Values.uibackend.serviceAccount.name }}
{{- end -}}
{{- end -}}
