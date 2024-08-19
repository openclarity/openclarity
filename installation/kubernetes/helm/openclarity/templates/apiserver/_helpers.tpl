{{/*
Base name for the apiserver components
*/}}
{{- define "openclarity.apiserver.name" -}}
{{- printf "%s-apiserver" (include  "openclarity.names.fullname" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Kubernetes standard labels
*/}}
{{- define "openclarity.apiserver.labels.standard" -}}
{{ include "openclarity.labels.standard" . }}
app.kubernetes.io/component: apiserver
{{- end -}}

{{/*
Labels to use on deploy.spec.selector.matchLabels and svc.spec.selector
*/}}
{{- define "openclarity.apiserver.labels.matchLabels" -}}
{{ include "openclarity.labels.matchLabels" . }}
app.kubernetes.io/component: apiserver
{{- end -}}

{{/*
Create the name of the service account to use
*/}}
{{- define "openclarity.apiserver.serviceAccountName" -}}
{{- if .Values.apiserver.serviceAccount.create -}}
    {{ default (include "openclarity.apiserver.name" .) .Values.apiserver.serviceAccount.name }}
{{- else -}}
    {{ default "default" .Values.apiserver.serviceAccount.name }}
{{- end -}}
{{- end -}}
