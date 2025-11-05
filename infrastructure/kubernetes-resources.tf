resource "kubernetes_namespace" "app" {
  count = var.kubernetes_resources_enabled ? 1 : 0

  metadata {
    name = "app"
    labels = {
      name        = "app"
      environment = var.environment
    }
  }
}

resource "random_password" "postgres_password" {
  count = var.kubernetes_resources_enabled ? 1 : 0

  length  = 32
  special = true
}

resource "random_password" "rabbitmq_password" {
  count = var.kubernetes_resources_enabled ? 1 : 0

  length  = 32
  special = true
}

resource "random_password" "keycloak_client_secret" {
  count = var.kubernetes_resources_enabled ? 1 : 0

  length  = 64
  special = true
}

resource "random_password" "keycloak_admin_password" {
  count = var.kubernetes_resources_enabled ? 1 : 0

  length  = 32
  special = true
}

resource "kubernetes_secret" "postgres_secret" {
  count = var.kubernetes_resources_enabled ? 1 : 0

  metadata {
    name      = "postgres-secret"
    namespace = kubernetes_namespace.app[0].metadata[0].name
  }

  data = {
    password = random_password.postgres_password[0].result
  }

  type = "Opaque"
}

resource "kubernetes_secret" "rabbitmq_secret" {
  count = var.kubernetes_resources_enabled ? 1 : 0

  metadata {
    name      = "rabbitmq-secret"
    namespace = kubernetes_namespace.app[0].metadata[0].name
  }

  data = {
    password = random_password.rabbitmq_password[0].result
  }

  type = "Opaque"
}

resource "kubernetes_secret" "keycloak_secret" {
  count = var.kubernetes_resources_enabled ? 1 : 0

  metadata {
    name      = "keycloak-secret"
    namespace = kubernetes_namespace.app[0].metadata[0].name
  }

  data = {
    admin-password = random_password.keycloak_admin_password[0].result
  }

  type = "Opaque"
}

resource "kubernetes_secret" "app_secrets" {
  count = var.kubernetes_resources_enabled ? 1 : 0

  metadata {
    name      = "app-secrets"
    namespace = kubernetes_namespace.app[0].metadata[0].name
  }

  data = {
    database-url           = "postgresql://app_user:${urlencode(random_password.postgres_password[0].result)}@postgres:5432/powerlifting_app"
    rabbitmq-url           = "amqp://admin:${urlencode(random_password.rabbitmq_password[0].result)}@rabbitmq:5672/"
    keycloak-client-secret = random_password.keycloak_client_secret[0].result
    spaces-key             = digitalocean_spaces_key.default.access_key
    spaces-secret          = digitalocean_spaces_key.default.secret_key
  }

  type = "Opaque"
}
