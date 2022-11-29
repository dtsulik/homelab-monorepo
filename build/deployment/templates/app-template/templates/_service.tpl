{{/*
Create service.
*/}}
{{- define "app-template.service" -}}
apiVersion: v1
kind: Service
metadata:
  name: {{ include "app-template.fullname" . }}
  labels:
    {{- include "app-template.labels" . | nindent 4 }}
spec:
  type: {{ .Values.service.type }}
  ports:
    - port: {{ .Values.service.port }}
      targetPort: http
      protocol: TCP
      name: http
  selector:
    {{- include "app-template.selectorLabels" . | nindent 4 }}
{{- end }}
