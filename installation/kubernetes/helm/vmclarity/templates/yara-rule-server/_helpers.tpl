{{/*
Base name for the yaraRuleServer components
*/}}
{{- define "vmclarity.yaraRuleServer.name" -}}
{{- printf "%s-yara-rule-server" (include  "vmclarity.names.fullname" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Kubernetes standard labels
*/}}
{{- define "vmclarity.yaraRuleServer.labels.standard" -}}
{{ include "vmclarity.labels.standard" . }}
app.kubernetes.io/component: yara-rule-server
{{- end -}}

{{/*
Labels to use on deploy.spec.selector.matchLabels and svc.spec.selector
*/}}
{{- define "vmclarity.yaraRuleServer.labels.matchLabels" -}}
{{ include "vmclarity.labels.matchLabels" . }}
app.kubernetes.io/component: yara-rule-server
{{- end -}}

{{/*
Create the name of the service account to use
*/}}
{{- define "vmclarity.yaraRuleServer.serviceAccountName" -}}
{{- if .Values.yaraRuleServer.serviceAccount.create -}}
    {{ default (include "vmclarity.yaraRuleServer.name" .) .Values.yaraRuleServer.serviceAccount.name }}
{{- else -}}
    {{ default "default" .Values.yaraRuleServer.serviceAccount.name }}
{{- end -}}
{{- end -}}
