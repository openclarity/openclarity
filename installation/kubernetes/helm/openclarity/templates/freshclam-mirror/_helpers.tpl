{{/*
Base name for the freshclamMirror components
*/}}
{{- define "openclarity.freshclamMirror.name" -}}
{{- printf "%s-freshclam-mirror" (include  "openclarity.names.fullname" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Kubernetes standard labels
*/}}
{{- define "openclarity.freshclamMirror.labels.standard" -}}
{{ include "openclarity.labels.standard" . }}
app.kubernetes.io/component: freshclam-mirror
{{- end -}}

{{/*
Labels to use on deploy.spec.selector.matchLabels and svc.spec.selector
*/}}
{{- define "openclarity.freshclamMirror.labels.matchLabels" -}}
{{ include "openclarity.labels.matchLabels" . }}
app.kubernetes.io/component: freshclam-mirror
{{- end -}}

{{/*
Create the name of the service account to use
*/}}
{{- define "openclarity.freshclamMirror.serviceAccountName" -}}
{{- if .Values.freshclamMirror.serviceAccount.create -}}
    {{ default (include "openclarity.freshclamMirror.name" .) .Values.freshclamMirror.serviceAccount.name }}
{{- else -}}
    {{ default "default" .Values.freshclamMirror.serviceAccount.name }}
{{- end -}}
{{- end -}}
