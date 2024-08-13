{{/*
Base name for the yaraRuleServer components
*/}}
{{- define "openclarity.yaraRuleServer.name" -}}
{{- printf "%s-yara-rule-server" (include  "openclarity.names.fullname" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Kubernetes standard labels
*/}}
{{- define "openclarity.yaraRuleServer.labels.standard" -}}
{{ include "openclarity.labels.standard" . }}
app.kubernetes.io/component: yara-rule-server
{{- end -}}

{{/*
Labels to use on deploy.spec.selector.matchLabels and svc.spec.selector
*/}}
{{- define "openclarity.yaraRuleServer.labels.matchLabels" -}}
{{ include "openclarity.labels.matchLabels" . }}
app.kubernetes.io/component: yara-rule-server
{{- end -}}

{{/*
Create the name of the service account to use
*/}}
{{- define "openclarity.yaraRuleServer.serviceAccountName" -}}
{{- if .Values.yaraRuleServer.serviceAccount.create -}}
    {{ default (include "openclarity.yaraRuleServer.name" .) .Values.yaraRuleServer.serviceAccount.name }}
{{- else -}}
    {{ default "default" .Values.yaraRuleServer.serviceAccount.name }}
{{- end -}}
{{- end -}}
