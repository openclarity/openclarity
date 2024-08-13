{{/*
Base name for the ui components
*/}}
{{- define "openclarity.ui.name" -}}
{{- printf "%s-ui" (include  "openclarity.names.fullname" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Kubernetes standard labels
*/}}
{{- define "openclarity.ui.labels.standard" -}}
{{ include "openclarity.labels.standard" . }}
app.kubernetes.io/component: ui
{{- end -}}

{{/*
Labels to use on deploy.spec.selector.matchLabels and svc.spec.selector
*/}}
{{- define "openclarity.ui.labels.matchLabels" -}}
{{ include "openclarity.labels.matchLabels" . }}
app.kubernetes.io/component: ui
{{- end -}}

{{/*
Create the name of the service account to use
*/}}
{{- define "openclarity.ui.serviceAccountName" -}}
{{- if .Values.ui.serviceAccount.create -}}
    {{ default (include "openclarity.ui.name" .) .Values.ui.serviceAccount.name }}
{{- else -}}
    {{ default "default" .Values.ui.serviceAccount.name }}
{{- end -}}
{{- end -}}
