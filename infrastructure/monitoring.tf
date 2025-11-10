# Monitoring namespace and resources
resource "kubernetes_namespace" "monitoring" {
  count = var.kubernetes_resources_enabled ? 1 : 0

  metadata {
    name = "monitoring"
    labels = {
      name        = "monitoring"
      environment = var.environment
    }
  }
}

# Grafana admin password
resource "random_password" "grafana_admin_password" {
  count = var.kubernetes_resources_enabled ? 1 : 0

  length  = 32
  special = true
}

# Update Grafana secret with random password
resource "kubernetes_secret" "grafana_secrets" {
  count = var.kubernetes_resources_enabled ? 1 : 0

  metadata {
    name      = "grafana-secrets"
    namespace = kubernetes_namespace.monitoring[0].metadata[0].name
  }

  data = {
    admin-password = random_password.grafana_admin_password[0].result
  }

  type = "Opaque"
}

# Ingress for Grafana
resource "kubernetes_ingress_v1" "grafana" {
  count = var.kubernetes_resources_enabled ? 1 : 0

  metadata {
    name      = "grafana-ingress"
    namespace = kubernetes_namespace.monitoring[0].metadata[0].name
    annotations = {
      "kubernetes.io/ingress.class"              = "nginx"
      "nginx.ingress.kubernetes.io/ssl-redirect" = "false"
    }
  }

  spec {
    rule {
      host = "grafana.${local.lb_ip}.nip.io"
      http {
        path {
          path      = "/"
          path_type = "Prefix"
          backend {
            service {
              name = "grafana"
              port {
                number = 3000
              }
            }
          }
        }
      }
    }
  }

  depends_on = [
    kubernetes_namespace.monitoring,
    data.kubernetes_service.nginx_ingress
  ]
}

# Ingress for Prometheus
resource "kubernetes_ingress_v1" "prometheus" {
  count = var.kubernetes_resources_enabled ? 1 : 0

  metadata {
    name      = "prometheus-ingress"
    namespace = kubernetes_namespace.monitoring[0].metadata[0].name
    annotations = {
      "kubernetes.io/ingress.class"              = "nginx"
      "nginx.ingress.kubernetes.io/ssl-redirect" = "false"
    }
  }

  spec {
    rule {
      host = "prometheus.${local.lb_ip}.nip.io"
      http {
        path {
          path      = "/"
          path_type = "Prefix"
          backend {
            service {
              name = "prometheus"
              port {
                number = 9090
              }
            }
          }
        }
      }
    }
  }

  depends_on = [
    kubernetes_namespace.monitoring,
    data.kubernetes_service.nginx_ingress
  ]
}

# Ingress for Loki
resource "kubernetes_ingress_v1" "loki" {
  count = var.kubernetes_resources_enabled ? 1 : 0

  metadata {
    name      = "loki-ingress"
    namespace = kubernetes_namespace.monitoring[0].metadata[0].name
    annotations = {
      "kubernetes.io/ingress.class"              = "nginx"
      "nginx.ingress.kubernetes.io/ssl-redirect" = "false"
    }
  }

  spec {
    rule {
      host = "loki.${local.lb_ip}.nip.io"
      http {
        path {
          path      = "/"
          path_type = "Prefix"
          backend {
            service {
              name = "loki"
              port {
                number = 3100
              }
            }
          }
        }
      }
    }
  }

  depends_on = [
    kubernetes_namespace.monitoring,
    data.kubernetes_service.nginx_ingress
  ]
}

# Ingress for RabbitMQ Management Console
resource "kubernetes_ingress_v1" "rabbitmq_management" {
  count = var.kubernetes_resources_enabled ? 1 : 0

  metadata {
    name      = "rabbitmq-management-ingress"
    namespace = kubernetes_namespace.app[0].metadata[0].name
    annotations = {
      "kubernetes.io/ingress.class"              = "nginx"
      "nginx.ingress.kubernetes.io/ssl-redirect" = "false"
    }
  }

  spec {
    rule {
      host = "rabbitmq.${local.lb_ip}.nip.io"
      http {
        path {
          path      = "/"
          path_type = "Prefix"
          backend {
            service {
              name = "rabbitmq"
              port {
                number = 15672
              }
            }
          }
        }
      }
    }
  }

  depends_on = [
    kubernetes_namespace.app,
    data.kubernetes_service.nginx_ingress
  ]
}
