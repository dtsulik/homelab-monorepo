{{/*
Create deployment.
*/}}
{{- define "app-template.deployment" -}}
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ include "app-template.fullname" . }}
  annotations:
    {{- if .Values.argocdAutoupdate.enabled }}
    argocd-image-updater.argoproj.io/image-list: "{{ .Values.image.repository }}:{{ .Values.argocdAutoupdate.tag }}"
    {{- end }}
  labels:
    {{- include "app-template.labels" . | nindent 4 }}
spec:
  revisionHistoryLimit: 1
  {{- if not .Values.autoscaling.enabled }}
  replicas: {{ .Values.replicaCount }}
  {{- end }}
  selector:
    matchLabels:
      {{- include "app-template.selectorLabels" . | nindent 6 }}
  template:
    metadata:
      {{- with .Values.podAnnotations }}
      annotations:
        {{- toYaml . | nindent 8 }}
      {{- end }}
      labels:
        {{- include "app-template.selectorLabels" . | nindent 8 }}
    spec:
      {{- with .Values.imagePullSecrets }}
      imagePullSecrets:
        {{- toYaml . | nindent 8 }}
      {{- end }}
      serviceAccountName: {{ include "app-template.serviceAccountName" . }}
      securityContext:
        {{- toYaml .Values.podSecurityContext | nindent 8 }}
      containers:
        - name: {{ .Chart.Name }}
          securityContext:
            {{- toYaml .Values.securityContext | nindent 12 }}
          image: "{{ .Values.image.repository }}:{{ .Values.image.tag | default .Chart.AppVersion }}"
          imagePullPolicy: {{ .Values.image.pullPolicy }}
          env:
          - name: K8S_SERVICE_PORT
            value: {{ .Values.service.port | quote }}
          {{- range $k, $v := .Values.env }}
          - name: {{ $k }}
            value: {{ $v | quote }}
          {{- end }}
          ports:
            - name: http
              containerPort: {{ .Values.service.port }}
              protocol: TCP
          livenessProbe:
            httpGet:
              path: /livez
              port: http
          readinessProbe:
            httpGet:
              path: /readyz
              port: http
          resources:
            {{- toYaml .Values.resources | nindent 12 }}
      {{- with .Values.nodeSelector }}
      nodeSelector:
        {{- toYaml . | nindent 8 }}
      {{- end }}
      {{- with .Values.affinity }}
      affinity:
        {{- toYaml . | nindent 8 }}
      {{- end }}
      {{- with .Values.tolerations }}
      tolerations:
        {{- toYaml . | nindent 8 }}
      {{- end }}
{{- end }}
