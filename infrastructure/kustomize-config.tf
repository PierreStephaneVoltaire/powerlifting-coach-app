locals {
  lb_ip = data.kubernetes_service.nginx_ingress.status[0].load_balancer[0].ingress[0].ip
}

resource "local_file" "kustomization_patches" {
  filename = "${path.module}/../k8s/overlays/production/kustomization.yaml"
  content  = <<-EOT
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

namespace: app

resources:
  - ../../base

patchesStrategicMerge:
  - production-patches.yaml
  - frontend-patch.yaml
  - auth-service-patch.yaml

patchesJson6902:
  - target:
      group: networking.k8s.io
      version: v1
      kind: Ingress
      name: powerlifting-coach-ingress
    patch: |-
      - op: replace
        path: /spec/rules/0/host
        value: app.${local.lb_ip}.nip.io
      - op: replace
        path: /spec/rules/1/host
        value: api.${local.lb_ip}.nip.io
      - op: replace
        path: /spec/rules/2/host
        value: auth.${local.lb_ip}.nip.io

images:
  - name: auth-service
    newName: ghcr.io/pierrestephanevoltaire/powerlifting-coach/auth-service
    newTag: latest
  - name: user-service
    newName: ghcr.io/pierrestephanevoltaire/powerlifting-coach/user-service
    newTag: latest
  - name: video-service
    newName: ghcr.io/pierrestephanevoltaire/powerlifting-coach/video-service
    newTag: latest
  - name: settings-service
    newName: ghcr.io/pierrestephanevoltaire/powerlifting-coach/settings-service
    newTag: latest
  - name: program-service
    newName: ghcr.io/pierrestephanevoltaire/powerlifting-coach/program-service
    newTag: latest
  - name: coach-service
    newName: ghcr.io/pierrestephanevoltaire/powerlifting-coach/coach-service
    newTag: latest
  - name: notification-service
    newName: ghcr.io/pierrestephanevoltaire/powerlifting-coach/notification-service
    newTag: latest
  - name: frontend
    newName: ghcr.io/pierrestephanevoltaire/powerlifting-coach/frontend
    newTag: latest

commonLabels:
  environment: production

replicas:
  - name: postgres
    count: 1
  - name: valkey
    count: 1
  - name: rabbitmq
    count: 1
EOT

  depends_on = [data.kubernetes_service.nginx_ingress]
}

resource "local_file" "frontend_patch" {
  filename = "${path.module}/../k8s/overlays/production/frontend-patch.yaml"
  content  = <<-EOT
apiVersion: apps/v1
kind: Deployment
metadata:
  name: frontend
  namespace: app
spec:
  template:
    spec:
      containers:
      - name: frontend
        env:
        - name: REACT_APP_API_URL
          value: "http://api.${local.lb_ip}.nip.io"
        - name: REACT_APP_AUTH_URL
          value: "http://api.${local.lb_ip}.nip.io/auth"
EOT

  depends_on = [data.kubernetes_service.nginx_ingress]
}

resource "local_file" "auth_service_patch" {
  filename = "${path.module}/../k8s/overlays/production/auth-service-patch.yaml"
  content  = <<-EOT
apiVersion: apps/v1
kind: Deployment
metadata:
  name: auth-service
  namespace: app
spec:
  template:
    spec:
      containers:
      - name: auth-service
        env:
        - name: KEYCLOAK_URL
          value: "http://auth.${local.lb_ip}.nip.io"
EOT

  depends_on = [data.kubernetes_service.nginx_ingress]
}
