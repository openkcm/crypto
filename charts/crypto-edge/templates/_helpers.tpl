{{/*
Expand the name of the chart.
*/}}
{{- define "crypto-edge.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "crypto-edge.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s" $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Namespace for all resources to be installed into
If not defined in values file then the helm release namespace is used
By default this is not set so the helm release namespace will be used

This gets around an problem within helm discussed here
https://github.com/helm/helm/issues/5358
*/}}
{{- define "crypto-edge.namespace" -}}
    {{ .Values.namespace | default .Release.Namespace }}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "crypto-edge.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "crypto-edge.labels" -}}
helm.sh/chart: {{ include "crypto-edge.chart" . }}
{{ include "crypto-edge.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Encrypto Selector labels
*/}}
{{- define "crypto-edge.encrypto.labels" -}}
{{ include "crypto-edge.labels" . }}
{{ include "crypto-edge.name" . }}.openkcm.io/component: encrypto
{{- end }}

{{/*
Tenant Manager Selector labels
*/}}
{{- define "crypto-edge.tenant-manager.labels" -}}
{{ include "crypto-edge.labels" . }}
{{ include "crypto-edge.name" . }}.openkcm.io/component: tenant-manager
{{- end }}


{{/*
Common Selector labels
*/}}
{{- define "crypto-edge.selectorLabels" -}}
app.kubernetes.io/name: {{ include "crypto-edge.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/component: {{ .Chart.Name }}
{{- end }}

{{/*
Encrypto Selector labels
*/}}
{{- define "crypto-edge.encrypto.selectorLabels" -}}
{{ include "crypto-edge.selectorLabels" . }}
{{ include "crypto-edge.name" . }}.openkcm.io/component: encrypto
{{- end }}

{{/*
Tenant Manager Selector labels
*/}}
{{- define "crypto-edge.tenant-manager.selectorLabels" -}}
{{ include "crypto-edge.selectorLabels" . }}
{{ include "crypto-edge.name" . }}.openkcm.io/component: tenant-manager
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "crypto-edge.serviceAccountName" -}}
{{- if .Values.common.serviceAccount.create }}
{{- default (include "crypto-edge.fullname" .) .Values.common.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.common.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Util function for generating the image URL based on the provided options.
*/}}
{{- define "crypto-edge.image" -}}
{{- $defaultTag := index . 1 -}}
{{- with index . 0 -}}
{{- if .registry -}}{{ printf "%s/%s" .registry .repository }}{{- else -}}{{- .repository -}}{{- end -}}
{{- if .digest -}}{{ printf "@%s" .digest }}{{- else -}}{{ printf ":%s" (default $defaultTag .tag) }}{{- end -}}
{{- end }}
{{- end }}
