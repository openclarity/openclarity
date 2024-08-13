{{/*
Base name for the crDiscoveryServer components
*/}}
{{- define "openclarity.crDiscoveryServer.name" -}}
{{- printf "%s-cr-discovery-server" (include  "openclarity.names.fullname" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Kubernetes standard labels
*/}}
{{- define "openclarity.crDiscoveryServer.labels.standard" -}}
{{ include "openclarity.labels.standard" . }}
app.kubernetes.io/component: cr-discovery-server
{{- end -}}

{{/*
Labels to use on deploy.spec.selector.matchLabels and svc.spec.selector
*/}}
{{- define "openclarity.crDiscoveryServer.labels.matchLabels" -}}
{{ include "openclarity.labels.matchLabels" . }}
app.kubernetes.io/component: cr-discovery-server
{{- end -}}

{{/*
Create the name of the service account to use
*/}}
{{- define "openclarity.crDiscoveryServer.serviceAccountName" -}}
{{- if .Values.crDiscoveryServer.serviceAccount.create -}}
    {{ default (include "openclarity.crDiscoveryServer.name" .) .Values.crDiscoveryServer.serviceAccount.name }}
{{- else -}}
    {{ default "default" .Values.crDiscoveryServer.serviceAccount.name }}
{{- end -}}
{{- end -}}
