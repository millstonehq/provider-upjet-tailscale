{{/*
Expand the name of the chart.
*/}}
{{- define "provider-tailscale.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "provider-tailscale.fullname" -}}
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
{{- define "provider-tailscale.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "provider-tailscale.labels" -}}
helm.sh/chart: {{ include "provider-tailscale.chart" . }}
{{ include "provider-tailscale.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- with .Values.additionalLabels }}
{{ toYaml . }}
{{- end }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "provider-tailscale.selectorLabels" -}}
app.kubernetes.io/name: {{ include "provider-tailscale.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Provider package image
*/}}
{{- define "provider-tailscale.image" -}}
{{- $tag := .Values.provider.package.tag | default .Chart.AppVersion }}
{{- printf "%s:%s" .Values.provider.package.repository $tag }}
{{- end }}

{{/*
Runtime config name
*/}}
{{- define "provider-tailscale.runtimeConfigName" -}}
{{- .Values.provider.runtimeConfig.name | default (printf "%s-config" (include "provider-tailscale.fullname" .)) }}
{{- end }}

{{/*
ProviderConfig name
*/}}
{{- define "provider-tailscale.providerConfigName" -}}
{{- .Values.providerConfig.name | default "default" }}
{{- end }}

{{/*
Secret name
*/}}
{{- define "provider-tailscale.secretName" -}}
{{- .Values.secret.name | default "tailscale-creds" }}
{{- end }}

{{/*
Common annotations (including ArgoCD sync wave if specified)
*/}}
{{- define "provider-tailscale.annotations" -}}
{{- $annotations := .annotations | default dict }}
{{- if .syncWave }}
argocd.argoproj.io/sync-wave: {{ .syncWave | quote }}
{{- end }}
{{- with $annotations }}
{{ toYaml . }}
{{- end }}
{{- with $.root.Values.additionalAnnotations }}
{{ toYaml . }}
{{- end }}
{{- end }}
