{{/*
Base name for the freshclamMirror components
*/}}
{{- define "vmclarity.freshclamMirror.name" -}}
{{- printf "%s-freshclam-mirror" (include  "vmclarity.names.fullname" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Kubernetes standard labels
*/}}
{{- define "vmclarity.freshclamMirror.labels.standard" -}}
{{ include "vmclarity.labels.standard" . }}
app.kubernetes.io/component: freshclam-mirror
{{- end -}}

{{/*
Labels to use on deploy.spec.selector.matchLabels and svc.spec.selector
*/}}
{{- define "vmclarity.freshclamMirror.labels.matchLabels" -}}
{{ include "vmclarity.labels.matchLabels" . }}
app.kubernetes.io/component: freshclam-mirror
{{- end -}}

{{/*
Create the name of the service account to use
*/}}
{{- define "vmclarity.freshclamMirror.serviceAccountName" -}}
{{- if .Values.freshclamMirror.serviceAccount.create -}}
    {{ default (include "vmclarity.freshclamMirror.name" .) .Values.freshclamMirror.serviceAccount.name }}
{{- else -}}
    {{ default "default" .Values.freshclamMirror.serviceAccount.name }}
{{- end -}}
{{- end -}}
