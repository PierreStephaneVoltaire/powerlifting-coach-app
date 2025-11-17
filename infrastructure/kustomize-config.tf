resource "local_file" "kustomization_patches" {
  count = var.kubernetes_resources_enabled && !var.stopped ? 1 : 0

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
  - video-service-patch.yaml
  - media-processor-service-patch.yaml

patchesJson6902:
  - target:
      group: gateway.networking.k8s.io
      version: v1
      kind: HTTPRoute
      name: frontend-route
    patch: |-
      - op: replace
        path: /spec/hostnames/0
        value: app.${var.domain_name}
  - target:
      group: gateway.networking.k8s.io
      version: v1
      kind: HTTPRoute
      name: api-route
    patch: |-
      - op: replace
        path: /spec/hostnames/0
        value: api.${var.domain_name}
  - target:
      group: gateway.networking.k8s.io
      version: v1
      kind: HTTPRoute
      name: auth-route
    patch: |-
      - op: replace
        path: /spec/hostnames/0
        value: auth.${var.domain_name}
  - target:
      group: cert-manager.io
      version: v1
      kind: Certificate
      name: app-tls
    patch: |-
      - op: replace
        path: /spec/dnsNames/0
        value: app.${var.domain_name}
      - op: replace
        path: /spec/dnsNames/1
        value: api.${var.domain_name}
      - op: replace
        path: /spec/dnsNames/2
        value: auth.${var.domain_name}

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
}

resource "local_file" "frontend_patch" {
  count = var.kubernetes_resources_enabled && !var.stopped ? 1 : 0

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
          value: "https://api.${var.domain_name}"
        - name: REACT_APP_AUTH_URL
          value: "https://api.${var.domain_name}/auth"
EOT
}

resource "local_file" "auth_service_patch" {
  count = var.kubernetes_resources_enabled && !var.stopped ? 1 : 0

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
          value: "http://keycloak:8080"
EOT
}

resource "local_file" "video_service_patch" {
  count = var.kubernetes_resources_enabled && !var.stopped ? 1 : 0

  filename = "${path.module}/../k8s/overlays/production/video-service-patch.yaml"
  content  = <<-EOT
apiVersion: apps/v1
kind: Deployment
metadata:
  name: video-service
  namespace: app
spec:
  template:
    spec:
      containers:
      - name: video-service
        env:
        - name: SPACES_ENDPOINT
          value: "https://${aws_s3_bucket.videos.bucket_regional_domain_name}"
        - name: SPACES_BUCKET
          value: "${aws_s3_bucket.videos.id}"
        - name: SPACES_REGION
          value: "${var.aws_region}"
EOT
}

resource "local_file" "media_processor_service_patch" {
  count = var.kubernetes_resources_enabled && !var.stopped ? 1 : 0

  filename = "${path.module}/../k8s/overlays/production/media-processor-service-patch.yaml"
  content  = <<-EOT
apiVersion: apps/v1
kind: Deployment
metadata:
  name: media-processor-service
  namespace: app
spec:
  template:
    spec:
      containers:
      - name: media-processor
        env:
        - name: SPACES_ENDPOINT
          value: "https://${aws_s3_bucket.videos.bucket_regional_domain_name}"
        - name: SPACES_BUCKET
          value: "${aws_s3_bucket.videos.id}"
        - name: SPACES_REGION
          value: "${var.aws_region}"
EOT
}
