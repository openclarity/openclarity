{{/*
Base name for the crDiscoveryServer components
*/}}
{{- define "vmclarity.crDiscoveryServer.name" -}}
{{- printf "%s-cr-discovery-server" (include  "vmclarity.names.fullname" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Kubernetes standard labels
*/}}
{{- define "vmclarity.crDiscoveryServer.labels.standard" -}}
{{ include "vmclarity.labels.standard" . }}
app.kubernetes.io/component: cr-discovery-server
{{- end -}}

{{/*
Labels to use on deploy.spec.selector.matchLabels and svc.spec.selector
*/}}
{{- define "vmclarity.crDiscoveryServer.labels.matchLabels" -}}
{{ include "vmclarity.labels.matchLabels" . }}
app.kubernetes.io/component: cr-discovery-server
{{- end -}}

{{/*
Create the name of the service account to use
*/}}
{{- define "vmclarity.crDiscoveryServer.serviceAccountName" -}}
{{- if .Values.crDiscoveryServer.serviceAccount.create -}}
    {{ default (include "vmclarity.crDiscoveryServer.name" .) .Values.crDiscoveryServer.serviceAccount.name }}
{{- else -}}
    {{ default "default" .Values.crDiscoveryServer.serviceAccount.name }}
{{- end -}}
{{- end -}}
