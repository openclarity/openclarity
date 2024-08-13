{{/*
Base name for the swaggerUI components
*/}}
{{- define "openclarity.swaggerUI.name" -}}
{{- printf "%s-swagger-ui" (include  "openclarity.names.fullname" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Kubernetes standard labels
*/}}
{{- define "openclarity.swaggerUI.labels.standard" -}}
{{ include "openclarity.labels.standard" . }}
app.kubernetes.io/component: swagger-ui
{{- end -}}

{{/*
Labels to use on deploy.spec.selector.matchLabels and svc.spec.selector
*/}}
{{- define "openclarity.swaggerUI.labels.matchLabels" -}}
{{ include "openclarity.labels.matchLabels" . }}
app.kubernetes.io/component: swagger-ui
{{- end -}}

{{/*
Create the name of the service account to use
*/}}
{{- define "openclarity.swaggerUI.serviceAccountName" -}}
{{- if .Values.swaggerUI.serviceAccount.create -}}
    {{ default (include "openclarity.swaggerUI.name" .) .Values.swaggerUI.serviceAccount.name }}
{{- else -}}
    {{ default "default" .Values.swaggerUI.serviceAccount.name }}
{{- end -}}
{{- end -}}
