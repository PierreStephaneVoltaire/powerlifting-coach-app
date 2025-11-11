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
    database-url            = "postgresql://app_user:${urlencode(random_password.postgres_password[0].result)}@postgres:5432/powerlifting_app?sslmode=disable"
    rabbitmq-url            = "amqp://admin:${urlencode(random_password.rabbitmq_password[0].result)}@rabbitmq:5672/"
    keycloak-client-secret  = random_password.keycloak_client_secret[0].result
    keycloak-admin-password = random_password.keycloak_admin_password[0].result
    spaces-key              = azurerm_storage_account.videos.name
    spaces-secret           = azurerm_storage_account.videos.primary_access_key
  }

  type = "Opaque"
}

resource "kubernetes_secret" "azure_email_secret" {
  count = var.kubernetes_resources_enabled ? 1 : 0

  metadata {
    name      = "azure-email-secret"
    namespace = kubernetes_namespace.app[0].metadata[0].name
  }

  data = {
    smtp-host     = local.azure_email_smtp_host
    smtp-username = var.domain_name != "localhost" ? "noreply@${var.domain_name}" : ""
    smtp-password = var.domain_name != "localhost" ? azurerm_communication_service.this[0].primary_connection_string : ""
    from-email    = var.domain_name != "localhost" ? "noreply@${var.domain_name}" : ""
  }

  type = "Opaque"

  depends_on = [
    azurerm_communication_service.this
  ]
}

resource "kubernetes_secret" "google_oauth_secret" {
  count = var.kubernetes_resources_enabled ? 1 : 0

  metadata {
    name      = "google-oauth-secret"
    namespace = kubernetes_namespace.app[0].metadata[0].name
  }

  data = {
    client-id     = var.google_oauth_client_id
    client-secret = var.google_oauth_client_secret
  }

  type = "Opaque"
}
