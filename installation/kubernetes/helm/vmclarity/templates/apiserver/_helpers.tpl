{{/*
Base name for the apiserver components
*/}}
{{- define "vmclarity.apiserver.name" -}}
{{- printf "%s-apiserver" (include  "vmclarity.names.fullname" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Kubernetes standard labels
*/}}
{{- define "vmclarity.apiserver.labels.standard" -}}
{{ include "vmclarity.labels.standard" . }}
app.kubernetes.io/component: apiserver
{{- end -}}

{{/*
Labels to use on deploy.spec.selector.matchLabels and svc.spec.selector
*/}}
{{- define "vmclarity.apiserver.labels.matchLabels" -}}
{{ include "vmclarity.labels.matchLabels" . }}
app.kubernetes.io/component: apiserver
{{- end -}}

{{/*
Create the name of the service account to use
*/}}
{{- define "vmclarity.apiserver.serviceAccountName" -}}
{{- if .Values.apiserver.serviceAccount.create -}}
    {{ default (include "vmclarity.apiserver.name" .) .Values.apiserver.serviceAccount.name }}
{{- else -}}
    {{ default "default" .Values.apiserver.serviceAccount.name }}
{{- end -}}
{{- end -}}
