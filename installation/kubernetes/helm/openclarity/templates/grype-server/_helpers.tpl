{{/*
Base name for the grypeServer components
*/}}
{{- define "openclarity.grypeServer.name" -}}
{{- printf "%s-grype-server" (include  "openclarity.names.fullname" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Kubernetes standard labels
*/}}
{{- define "openclarity.grypeServer.labels.standard" -}}
{{ include "openclarity.labels.standard" . }}
app.kubernetes.io/component: grype-server
{{- end -}}

{{/*
Labels to use on deploy.spec.selector.matchLabels and svc.spec.selector
*/}}
{{- define "openclarity.grypeServer.labels.matchLabels" -}}
{{ include "openclarity.labels.matchLabels" . }}
app.kubernetes.io/component: grype-server
{{- end -}}

{{/*
Create the name of the service account to use
*/}}
{{- define "openclarity.grypeServer.serviceAccountName" -}}
{{- if .Values.grypeServer.serviceAccount.create -}}
    {{ default (include "openclarity.grypeServer.name" .) .Values.grypeServer.serviceAccount.name }}
{{- else -}}
    {{ default "default" .Values.grypeServer.serviceAccount.name }}
{{- end -}}
{{- end -}}
