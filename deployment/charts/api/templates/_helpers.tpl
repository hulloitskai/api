{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "chart.name" -}}
	{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this
(by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "chart.fullname" -}}
	{{- if .Values.fullnameOverride -}}
		{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
	{{- else -}}
		{{- $name := default .Chart.Name .Values.nameOverride -}}
		{{- if contains $name .Release.Name -}}
			{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
		{{- else -}}
			{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
		{{- end -}}
	{{- end -}}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "chart.chart" -}}
	{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create the name for the Varnish cache proxy.
*/}}
{{- define "varnish.name" -}}
	{{- $name := include "chart.name" . -}}
	{{- printf "%s-varnish" $name -}}
{{- end -}}
{{- define "varnish.fullname" -}}
	{{- $fullName := include "chart.fullname" . -}}
	{{- printf "%s-varnish" $fullName -}}
{{- end -}}

{{/*
Create the name for the frontend components.
*/}}
{{- define "frontend.name" -}}
	{{- $name := include "chart.name" . -}}
	{{- printf "%s-frontend" $name -}}
{{- end -}}
{{- define "frontend.fullname" -}}
	{{- $fullName := include "chart.fullname" . -}}
	{{- printf "%s-frontend" $fullName -}}
{{- end -}}
