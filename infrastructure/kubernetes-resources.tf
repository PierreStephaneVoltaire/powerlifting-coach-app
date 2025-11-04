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
    database-url           = "postgresql://app_user:${random_password.postgres_password.result}@postgres:5432/powerlifting_app"
    rabbitmq-url           = "amqp://admin:${random_password.rabbitmq_password.result}@rabbitmq:5672/"
    keycloak-client-secret = random_password.keycloak_client_secret.result
    spaces-key             = digitalocean_spaces_key.default.access_key
    spaces-secret          = digitalocean_spaces_key.default.secret_key
  }

  type = "Opaque"
}
