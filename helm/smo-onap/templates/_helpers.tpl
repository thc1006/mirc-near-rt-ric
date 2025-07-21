{{/*
Expand the name of the chart.
*/}}
{{- define "smo-onap.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "smo-onap.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "smo-onap.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "smo-onap.labels" -}}
helm.sh/chart: {{ include "smo-onap.chart" . }}
{{ include "smo-onap.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: o-ran-smo
{{- end }}

{{/*
Selector labels
*/}}
{{- define "smo-onap.selectorLabels" -}}
app.kubernetes.io/name: {{ include "smo-onap.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "smo-onap.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "smo-onap.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Common ONAP environment variables
*/}}
{{- define "smo-onap.commonEnv" -}}
- name: NAMESPACE
  valueFrom:
    fieldRef:
      fieldPath: metadata.namespace
- name: RELEASE_NAME
  value: {{ .Release.Name }}
- name: CHART_VERSION
  value: {{ .Chart.Version }}
{{- end }}

{{/*
Resource limits helper
*/}}
{{- define "smo-onap.resources" -}}
{{- if .resources }}
resources:
  {{- if .resources.limits }}
  limits:
    {{- if .resources.limits.cpu }}
    cpu: {{ .resources.limits.cpu }}
    {{- end }}
    {{- if .resources.limits.memory }}
    memory: {{ .resources.limits.memory }}
    {{- end }}
  {{- end }}
  {{- if .resources.requests }}
  requests:
    {{- if .resources.requests.cpu }}
    cpu: {{ .resources.requests.cpu }}
    {{- end }}
    {{- if .resources.requests.memory }}
    memory: {{ .resources.requests.memory }}
    {{- end }}
  {{- end }}
{{- end }}
{{- end }}

{{/*
O-RAN interface endpoints
*/}}
{{- define "smo-onap.a1Endpoint" -}}
{{- printf "http://%s-a1mediator:8080" .Release.Name }}
{{- end }}

{{- define "smo-onap.o1Endpoint" -}}
{{- printf "http://%s-o1mediator:8080" .Release.Name }}
{{- end }}

{{- define "smo-onap.e2Endpoint" -}}
{{- printf "%s-e2term:38000" .Release.Name }}
{{- end }}

{{/*
Database connection strings
*/}}
{{- define "smo-onap.dbHost" -}}
{{- printf "%s-mariadb-galera" .Release.Name }}
{{- end }}

{{- define "smo-onap.postgresHost" -}}
{{- printf "%s-postgresql" .Release.Name }}
{{- end }}

{{- define "smo-onap.redisHost" -}}
{{- printf "%s-redis-master" .Release.Name }}
{{- end }}

{{/*
SMO O2 Interface Configuration
*/}}
{{- define "smo-onap.o2Config" -}}
o2ims_endpoint: {{ printf "http://%s-smo:5005" .Release.Name }}
o2dms_endpoint: {{ printf "http://%s-smo:5006" .Release.Name }}
netconf_endpoint: {{ printf "%s-netconf:830" .Release.Name }}
restconf_endpoint: {{ printf "http://%s-restconf:8181" .Release.Name }}
{{- end }}