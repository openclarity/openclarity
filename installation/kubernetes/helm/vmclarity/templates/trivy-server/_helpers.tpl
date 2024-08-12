{{/*
Base name for the trivyServer components
*/}}
{{- define "vmclarity.trivyServer.name" -}}
{{- printf "%s-trivy-server" (include  "vmclarity.names.fullname" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Kubernetes standard labels
*/}}
{{- define "vmclarity.trivyServer.labels.standard" -}}
{{ include "vmclarity.labels.standard" . }}
app.kubernetes.io/component: trivy-server
{{- end -}}

{{/*
Labels to use on deploy.spec.selector.matchLabels and svc.spec.selector
*/}}
{{- define "vmclarity.trivyServer.labels.matchLabels" -}}
{{ include "vmclarity.labels.matchLabels" . }}
app.kubernetes.io/component: trivy-server
{{- end -}}

{{/*
Create the name of the service account to use
*/}}
{{- define "vmclarity.trivyServer.serviceAccountName" -}}
{{- if .Values.trivyServer.serviceAccount.create -}}
    {{ default (include "vmclarity.trivyServer.name" .) .Values.trivyServer.serviceAccount.name }}
{{- else -}}
    {{ default "default" .Values.trivyServer.serviceAccount.name }}
{{- end -}}
{{- end -}}
