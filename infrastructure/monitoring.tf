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

resource "random_password" "grafana_admin_password" {
  count = var.kubernetes_resources_enabled ? 1 : 0

  length  = 32
  special = true
}

resource "helm_release" "loki" {
  count = var.kubernetes_resources_enabled && !var.stopped ? 1 : 0

  name             = "loki"
  repository       = "https://grafana.github.io/helm-charts"
  chart            = "loki"
  namespace        = kubernetes_namespace.monitoring[0].metadata[0].name
  create_namespace = false
  version          = "5.43.3"

  values = [
    yamlencode({
      loki = {
        auth_enabled = false
        commonConfig = {
          replication_factor = 1
        }
        storage = {
          type = "filesystem"
        }
      }
      singleBinary = {
        replicas = 1
        persistence = {
          enabled      = true
          storageClass = "gp3"
          size         = "30Gi"
        }
        resources = {
          requests = {
            cpu    = "100m"
            memory = "256Mi"
          }
          limits = {
            cpu    = "200m"
            memory = "512Mi"
          }
        }
      }
      gateway = {
        enabled = false
      }
      test = {
        enabled = false
      }
      monitoring = {
        lokiCanary = {
          enabled = false
        }
      }
    })
  ]

  depends_on = [
    module.eks_blueprints_addons,
    kubernetes_namespace.monitoring
  ]
}

resource "helm_release" "promtail" {
  count = var.kubernetes_resources_enabled && !var.stopped ? 1 : 0

  name             = "promtail"
  repository       = "https://grafana.github.io/helm-charts"
  chart            = "promtail"
  namespace        = kubernetes_namespace.monitoring[0].metadata[0].name
  create_namespace = false
  version          = "6.15.5"

  values = [
    yamlencode({
      config = {
        clients = [
          {
            url = "http://loki:3100/loki/api/v1/push"
          }
        ]
      }
      resources = {
        requests = {
          cpu    = "50m"
          memory = "64Mi"
        }
        limits = {
          cpu    = "100m"
          memory = "128Mi"
        }
      }
    })
  ]

  depends_on = [helm_release.loki]
}

resource "kubernetes_ingress_v1" "grafana" {
  count = var.kubernetes_resources_enabled ? 1 : 0

  metadata {
    name      = "grafana-ingress"
    namespace = "kube-prometheus-stack"
    annotations = {
      "kubernetes.io/ingress.class"                    = "nginx"
      "cert-manager.io/cluster-issuer"                 = "letsencrypt-prod"
      "nginx.ingress.kubernetes.io/ssl-redirect"       = "true"
      "nginx.ingress.kubernetes.io/force-ssl-redirect" = "true"
    }
  }

  spec {
    tls {
      hosts       = ["grafana.${var.domain_name}"]
      secret_name = "grafana-tls"
    }

    rule {
      host = "grafana.${var.domain_name}"
      http {
        path {
          path      = "/"
          path_type = "Prefix"
          backend {
            service {
              name = "kube-prometheus-stack-grafana"
              port {
                number = 80
              }
            }
          }
        }
      }
    }
  }

  depends_on = [
    module.eks_blueprints_addons,
    helm_release.nginx_ingress
  ]
}

resource "kubernetes_ingress_v1" "prometheus" {
  count = var.kubernetes_resources_enabled ? 1 : 0

  metadata {
    name      = "prometheus-ingress"
    namespace = "kube-prometheus-stack"
    annotations = {
      "kubernetes.io/ingress.class"                    = "nginx"
      "cert-manager.io/cluster-issuer"                 = "letsencrypt-prod"
      "nginx.ingress.kubernetes.io/ssl-redirect"       = "true"
      "nginx.ingress.kubernetes.io/force-ssl-redirect" = "true"
    }
  }

  spec {
    tls {
      hosts       = ["prometheus.${var.domain_name}"]
      secret_name = "prometheus-tls"
    }

    rule {
      host = "prometheus.${var.domain_name}"
      http {
        path {
          path      = "/"
          path_type = "Prefix"
          backend {
            service {
              name = "kube-prometheus-stack-prometheus"
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
    module.eks_blueprints_addons,
    helm_release.nginx_ingress
  ]
}

resource "kubernetes_ingress_v1" "loki" {
  count = var.kubernetes_resources_enabled ? 1 : 0

  metadata {
    name      = "loki-ingress"
    namespace = kubernetes_namespace.monitoring[0].metadata[0].name
    annotations = {
      "kubernetes.io/ingress.class"                    = "nginx"
      "cert-manager.io/cluster-issuer"                 = "letsencrypt-prod"
      "nginx.ingress.kubernetes.io/ssl-redirect"       = "true"
      "nginx.ingress.kubernetes.io/force-ssl-redirect" = "true"
    }
  }

  spec {
    tls {
      hosts       = ["loki.${var.domain_name}"]
      secret_name = "loki-tls"
    }

    rule {
      host = "loki.${var.domain_name}"
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
    helm_release.loki,
    helm_release.nginx_ingress
  ]
}

resource "kubernetes_ingress_v1" "rabbitmq_management" {
  count = var.kubernetes_resources_enabled ? 1 : 0

  metadata {
    name      = "rabbitmq-management-ingress"
    namespace = kubernetes_namespace.app[0].metadata[0].name
    annotations = {
      "kubernetes.io/ingress.class"                    = "nginx"
      "cert-manager.io/cluster-issuer"                 = "letsencrypt-prod"
      "nginx.ingress.kubernetes.io/ssl-redirect"       = "true"
      "nginx.ingress.kubernetes.io/force-ssl-redirect" = "true"
    }
  }

  spec {
    tls {
      hosts       = ["rabbitmq.${var.domain_name}"]
      secret_name = "rabbitmq-tls"
    }

    rule {
      host = "rabbitmq.${var.domain_name}"
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
    helm_release.nginx_ingress
  ]
}
