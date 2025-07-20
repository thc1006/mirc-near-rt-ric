{{/*
Expand the name of the chart.
*/}}
{{- define "oran-nearrt-ric.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "oran-nearrt-ric.fullname" -}}
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
{{- define "oran-nearrt-ric.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "oran-nearrt-ric.labels" -}}
helm.sh/chart: {{ include "oran-nearrt-ric.chart" . }}
{{ include "oran-nearrt-ric.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "oran-nearrt-ric.selectorLabels" -}}
app.kubernetes.io/name: {{ include "oran-nearrt-ric.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "oran-nearrt-ric.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "oran-nearrt-ric.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Create the name of the namespace to use
*/}}
{{- define "oran-nearrt-ric.namespace" -}}
{{- default .Release.Namespace .Values.namespaceOverride }}
{{- end }}

{{/*
Common labels for main dashboard
*/}}
{{- define "oran-nearrt-ric.mainDashboard.labels" -}}
{{ include "oran-nearrt-ric.labels" . }}
app.kubernetes.io/component: main-dashboard
{{- end }}

{{/*
Selector labels for main dashboard
*/}}
{{- define "oran-nearrt-ric.mainDashboard.selectorLabels" -}}
{{ include "oran-nearrt-ric.selectorLabels" . }}
app.kubernetes.io/component: main-dashboard
{{- end }}

{{/*
Common labels for xApp dashboard
*/}}
{{- define "oran-nearrt-ric.xappDashboard.labels" -}}
{{ include "oran-nearrt-ric.labels" . }}
app.kubernetes.io/component: xapp-dashboard
{{- end }}

{{/*
Selector labels for xApp dashboard
*/}}
{{- define "oran-nearrt-ric.xappDashboard.selectorLabels" -}}
{{ include "oran-nearrt-ric.selectorLabels" . }}
app.kubernetes.io/component: xapp-dashboard
{{- end }}

{{/*
Common labels for federated learning coordinator
*/}}
{{- define "oran-nearrt-ric.federatedLearning.labels" -}}
{{ include "oran-nearrt-ric.labels" . }}
app.kubernetes.io/component: fl-coordinator
{{- end }}

{{/*
Selector labels for federated learning coordinator
*/}}
{{- define "oran-nearrt-ric.federatedLearning.selectorLabels" -}}
{{ include "oran-nearrt-ric.selectorLabels" . }}
app.kubernetes.io/component: fl-coordinator
{{- end }}

{{/*
Image pull secrets
*/}}
{{- define "oran-nearrt-ric.imagePullSecrets" -}}
{{- if .Values.global.imagePullSecrets }}
imagePullSecrets:
{{- range .Values.global.imagePullSecrets }}
  - name: {{ . }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Return the proper image name for main dashboard
*/}}
{{- define "oran-nearrt-ric.mainDashboard.image" -}}
{{- $registryName := .Values.global.imageRegistry -}}
{{- $repositoryName := .Values.mainDashboard.image.repository -}}
{{- $tag := .Values.mainDashboard.image.tag | toString -}}
{{- if $registryName }}
{{- printf "%s/%s:%s" $registryName $repositoryName $tag -}}
{{- else }}
{{- printf "%s:%s" $repositoryName $tag -}}
{{- end }}
{{- end }}

{{/*
Return the proper image name for xApp dashboard
*/}}
{{- define "oran-nearrt-ric.xappDashboard.image" -}}
{{- $registryName := .Values.global.imageRegistry -}}
{{- $repositoryName := .Values.xappDashboard.image.repository -}}
{{- $tag := .Values.xappDashboard.image.tag | toString -}}
{{- if $registryName }}
{{- printf "%s/%s:%s" $registryName $repositoryName $tag -}}
{{- else }}
{{- printf "%s:%s" $repositoryName $tag -}}
{{- end }}
{{- end }}

{{/*
Return the proper image name for federated learning coordinator
*/}}
{{- define "oran-nearrt-ric.federatedLearning.image" -}}
{{- $registryName := .Values.global.imageRegistry -}}
{{- $repositoryName := .Values.federatedLearning.image.repository -}}
{{- $tag := .Values.federatedLearning.image.tag | toString -}}
{{- if $registryName }}
{{- printf "%s/%s:%s" $registryName $repositoryName $tag -}}
{{- else }}
{{- printf "%s:%s" $repositoryName $tag -}}
{{- end }}
{{- end }}

{{/*
Return the proper Storage Class
*/}}
{{- define "oran-nearrt-ric.storageClass" -}}
{{- if .Values.global.storageClass -}}
{{- if (eq "-" .Values.global.storageClass) -}}
{{- printf "storageClassName: \"\"" -}}
{{- else }}
{{- printf "storageClassName: %s" .Values.global.storageClass -}}
{{- end -}}
{{- end -}}
{{- end -}}