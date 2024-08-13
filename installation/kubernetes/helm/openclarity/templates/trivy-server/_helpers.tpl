{{/*
Base name for the trivyServer components
*/}}
{{- define "openclarity.trivyServer.name" -}}
{{- printf "%s-trivy-server" (include  "openclarity.names.fullname" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Kubernetes standard labels
*/}}
{{- define "openclarity.trivyServer.labels.standard" -}}
{{ include "openclarity.labels.standard" . }}
app.kubernetes.io/component: trivy-server
{{- end -}}

{{/*
Labels to use on deploy.spec.selector.matchLabels and svc.spec.selector
*/}}
{{- define "openclarity.trivyServer.labels.matchLabels" -}}
{{ include "openclarity.labels.matchLabels" . }}
app.kubernetes.io/component: trivy-server
{{- end -}}

{{/*
Create the name of the service account to use
*/}}
{{- define "openclarity.trivyServer.serviceAccountName" -}}
{{- if .Values.trivyServer.serviceAccount.create -}}
    {{ default (include "openclarity.trivyServer.name" .) .Values.trivyServer.serviceAccount.name }}
{{- else -}}
    {{ default "default" .Values.trivyServer.serviceAccount.name }}
{{- end -}}
{{- end -}}
