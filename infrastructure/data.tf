resource "random_password" "valkey_password" {
  length  = 32
  special = true
}

resource "random_password" "postgres_password" {
  length  = 32
  special = false
}



resource "kubernetes_secret" "valkey" {
  metadata {
    name      = "valkey"
    namespace = kubernetes_namespace.data.metadata[0].name
  }
  data = {
    password = random_password.valkey_password.result
    host     = "valkey.${kubernetes_namespace.data.metadata[0].name}.svc.cluster.local"
    port     = "6379"
  }
  type = "Opaque"
}


resource "kubernetes_secret" "postgres" {
  metadata {
    name      = "postgres"
    namespace = kubernetes_namespace.data.metadata[0].name
  }
  data = {
    username = "postgres"
    password = random_password.postgres_password.result
    host     = "postgres.${kubernetes_namespace.data.metadata[0].name}.svc.cluster.local"
    port     = "5432"
    database = "app"
  }
  type = "Opaque"
}






resource "kubernetes_stateful_set" "postgres" {
  metadata {
    name      = "postgres"
    namespace = kubernetes_namespace.data.metadata[0].name
  }

  spec {
    service_name = "postgres"
    replicas     = 1

    selector {
      match_labels = {
        app = "postgres"
      }
    }

    template {
      metadata {
        labels = {
          app = "postgres"
        }
      }

      spec {
        container {
          name  = "postgres"
          image = "postgres:16-alpine"

          port {
            container_port = 5432
            name           = "postgres"
          }

          env {
            name = "POSTGRES_USER"
            value_from {
              secret_key_ref {
                name = kubernetes_secret.postgres.metadata[0].name
                key  = "username"
              }
            }
          }

          env {
            name = "POSTGRES_PASSWORD"
            value_from {
              secret_key_ref {
                name = kubernetes_secret.postgres.metadata[0].name
                key  = "password"
              }
            }
          }

          env {
            name  = "POSTGRES_DB"
            value = "app"
          }

          env {
            name  = "PGDATA"
            value = "/var/lib/postgresql/data/pgdata"
          }

          resources {
            requests = {
              memory = "256Mi"
              cpu    = "250m"
            }
            limits = {
              memory = "1Gi"
              cpu    = "1000m"
            }
          }

          volume_mount {
            name       = "postgres-data"
            mount_path = "/var/lib/postgresql/data"
          }

          liveness_probe {
            exec {
              command = ["pg_isready", "-U", "postgres"]
            }
            initial_delay_seconds = 30
            period_seconds        = 10
          }

          readiness_probe {
            exec {
              command = ["pg_isready", "-U", "postgres"]
            }
            initial_delay_seconds = 5
            period_seconds        = 10
          }
        }
      }
    }

    volume_claim_template {
      metadata {
        name = "postgres-data"
      }

      spec {
        access_modes = ["ReadWriteOnce"]
        resources {
          requests = {
            storage = "10Gi"
          }
        }
      }
    }
  }
}

resource "kubernetes_service" "postgres" {
  metadata {
    name      = "postgres"
    namespace = kubernetes_namespace.data.metadata[0].name
  }

  spec {
    selector = {
      app = "postgres"
    }

    port {
      port        = 5432
      target_port = 5432
      protocol    = "TCP"
    }

    type = "ClusterIP"
  }
}


resource "kubernetes_stateful_set" "valkey" {
  metadata {
    name      = "valkey"
    namespace = kubernetes_namespace.data.metadata[0].name
  }

  spec {
    service_name = "valkey"
    replicas     = 1

    selector {
      match_labels = {
        app = "valkey"
      }
    }

    template {
      metadata {
        labels = {
          app = "valkey"
        }
      }

      spec {
        container {
          name  = "valkey"
          image = "valkey/valkey:7.2-alpine"

          port {
            container_port = 6379
            name           = "valkey"
          }

          args = [
            "--requirepass",
            "$(VALKEY_PASSWORD)",
            "--appendonly",
            "yes"
          ]

          env {
            name = "VALKEY_PASSWORD"
            value_from {
              secret_key_ref {
                name = kubernetes_secret.valkey.metadata[0].name
                key  = "password"
              }
            }
          }

          volume_mount {
            name       = "valkey-data"
            mount_path = "/data"
          }

          resources {
            requests = {
              memory = "256Mi"
              cpu    = "250m"
            }
            limits = {
              memory = "512Mi"
              cpu    = "500m"
            }
          }
        }
      }
    }

    volume_claim_template {
      metadata {
        name = "valkey-data"
      }

      spec {
        access_modes = ["ReadWriteOnce"]
        resources {
          requests = {
            storage = "5Gi"
          }
        }
      }
    }
  }
}