resource "kubernetes_namespace" "app" {
  metadata {
    name = "app"
    labels = {
      name        = "app"
      environment = var.environment
    }
  }
}

resource "random_password" "postgres_password" {
  length  = 32
  special = true
}

resource "random_password" "rabbitmq_password" {
  length  = 32
  special = true
}

resource "random_password" "keycloak_client_secret" {
  length  = 64
  special = true
}

resource "random_password" "keycloak_admin_password" {
  length  = 32
  special = true
}

resource "random_password" "grafana_admin_password" {
  length  = 32
  special = true
}

resource "kubernetes_secret" "postgres_secret" {
  metadata {
    name      = "postgres-secret"
    namespace = kubernetes_namespace.app.metadata[0].name
  }

  data = {
    password = random_password.postgres_password.result
  }

  type = "Opaque"
}

resource "kubernetes_secret" "rabbitmq_secret" {
  metadata {
    name      = "rabbitmq-secret"
    namespace = kubernetes_namespace.app.metadata[0].name
  }

  data = {
    password = random_password.rabbitmq_password.result
  }

  type = "Opaque"
}

resource "kubernetes_secret" "keycloak_secret" {
  metadata {
    name      = "keycloak-secret"
    namespace = kubernetes_namespace.app.metadata[0].name
  }

  data = {
    admin-password = random_password.keycloak_admin_password.result
  }

  type = "Opaque"
}

resource "kubernetes_secret" "app_secrets" {
  metadata {
    name      = "app-secrets"
    namespace = kubernetes_namespace.app.metadata[0].name
  }

  data = {
    database-url            = "postgresql://app_user:${urlencode(random_password.postgres_password.result)}@postgres:5432/powerlifting_app?sslmode=disable"
    rabbitmq-url            = "amqp://admin:${urlencode(random_password.rabbitmq_password.result)}@rabbitmq:5672/"
    keycloak-client-secret  = random_password.keycloak_client_secret.result
    keycloak-admin-password = random_password.keycloak_admin_password.result
    spaces-key              = data.terraform_remote_state.base.outputs.s3_videos_access_key
    spaces-secret           = data.terraform_remote_state.base.outputs.s3_videos_secret_key
  }

  type = "Opaque"
}

resource "kubernetes_secret" "ses_email_secret" {
  metadata {
    name      = "ses-email-secret"
    namespace = kubernetes_namespace.app.metadata[0].name
  }

  data = {
    smtp-host     = local.ses_smtp_endpoint
    smtp-port     = "587"
    smtp-username = data.terraform_remote_state.base.outputs.ses_smtp_username
    smtp-password = data.terraform_remote_state.base.outputs.ses_smtp_password
    from-email    = "noreply@${var.domain_name}"
  }

  type = "Opaque"
}

resource "kubernetes_secret" "google_oauth_secret" {
  metadata {
    name      = "google-oauth-secret"
    namespace = kubernetes_namespace.app.metadata[0].name
  }

  data = {
    client-id     = var.google_oauth_client_id
    client-secret = var.google_oauth_client_secret
  }

  type = "Opaque"
}

resource "local_file" "kustomization_patches" {
  count = var.stopped ? 0 : 1

  filename = "${path.module}/../../../k8s/overlays/production/kustomization.yaml"
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
  count = var.stopped ? 0 : 1

  filename = "${path.module}/../../../k8s/overlays/production/frontend-patch.yaml"
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
        resources:
          requests:
            memory: "32Mi"
            cpu: "25m"
          limits:
            memory: "64Mi"
            cpu: "100m"
EOT
}

resource "local_file" "auth_service_patch" {
  count = var.stopped ? 0 : 1

  filename = "${path.module}/../../../k8s/overlays/production/auth-service-patch.yaml"
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
  count = var.stopped ? 0 : 1

  filename = "${path.module}/../../../k8s/overlays/production/video-service-patch.yaml"
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
          value: "https://${data.terraform_remote_state.base.outputs.s3_videos_bucket_domain}"
        - name: SPACES_BUCKET
          value: "${data.terraform_remote_state.base.outputs.s3_videos_bucket}"
        - name: SPACES_REGION
          value: "${var.aws_region}"
EOT
}

resource "local_file" "media_processor_service_patch" {
  count = var.stopped ? 0 : 1

  filename = "${path.module}/../../../k8s/overlays/production/media-processor-service-patch.yaml"
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
          value: "https://${data.terraform_remote_state.base.outputs.s3_videos_bucket_domain}"
        - name: SPACES_BUCKET
          value: "${data.terraform_remote_state.base.outputs.s3_videos_bucket}"
        - name: SPACES_REGION
          value: "${var.aws_region}"
EOT
}
