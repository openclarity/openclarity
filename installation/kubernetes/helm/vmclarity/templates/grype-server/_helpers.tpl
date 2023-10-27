{{/*
Base name for the grypeServer components
*/}}
{{- define "vmclarity.grypeServer.name" -}}
{{- printf "%s-grype-server" (include  "vmclarity.names.fullname" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Kubernetes standard labels
*/}}
{{- define "vmclarity.grypeServer.labels.standard" -}}
{{ include "vmclarity.labels.standard" . }}
app.kubernetes.io/component: grype-server
{{- end -}}

{{/*
Labels to use on deploy.spec.selector.matchLabels and svc.spec.selector
*/}}
{{- define "vmclarity.grypeServer.labels.matchLabels" -}}
{{ include "vmclarity.labels.matchLabels" . }}
app.kubernetes.io/component: grype-server
{{- end -}}

{{/*
Create the name of the service account to use
*/}}
{{- define "vmclarity.grypeServer.serviceAccountName" -}}
{{- if .Values.grypeServer.serviceAccount.create -}}
    {{ default (include "vmclarity.grypeServer.name" .) .Values.grypeServer.serviceAccount.name }}
{{- else -}}
    {{ default "default" .Values.grypeServer.serviceAccount.name }}
{{- end -}}
{{- end -}}
