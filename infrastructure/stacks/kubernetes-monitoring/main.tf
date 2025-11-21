resource "helm_release" "kube_prometheus_stack" {
  count = var.stopped ? 0 : 1

  name             = "prometheus"
  repository       = "https://prometheus-community.github.io/helm-charts"
  chart            = "kube-prometheus-stack"
  namespace        = "monitoring"
  create_namespace = true
  version          = "55.5.0"

  values = [
    yamlencode({
      prometheus = {
        prometheusSpec = {
          retention = "15d"
          resources = {
            requests = {
              cpu    = "200m"
              memory = "512Mi"
            }
            limits = {
              cpu    = "500m"
              memory = "2Gi"
            }
          }
          storageSpec = {
            volumeClaimTemplate = {
              spec = {
                storageClassName = "local-path"
                accessModes      = ["ReadWriteOnce"]
                resources = {
                  requests = {
                    storage = "50Gi"
                  }
                }
              }
            }
          }
        }
      }
      grafana = {
        enabled       = true
        adminPassword = data.terraform_remote_state.kubernetes_base.outputs.grafana_admin_password
        persistence = {
          enabled          = true
          storageClassName = "local-path"
          size             = "10Gi"
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
      alertmanager = {
        enabled = false
      }
    })
  ]
}

resource "helm_release" "loki" {
  count = var.stopped ? 0 : 1

  name             = "loki"
  repository       = "https://grafana.github.io/helm-charts"
  chart            = "loki"
  namespace        = "monitoring"
  create_namespace = true
  version          = "5.43.3"

  values = [
    yamlencode({
      loki = {
        auth_enabled = false
        commonConfig = {
          replication_factor = 1
        }
      }
      singleBinary = {
        replicas = 1
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
        persistence = {
          storageClass = "local-path"
          size         = "30Gi"
        }
      }
    })
  ]

  depends_on = [helm_release.kube_prometheus_stack]
}

resource "helm_release" "promtail" {
  count = var.stopped ? 0 : 1

  name       = "promtail"
  repository = "https://grafana.github.io/helm-charts"
  chart      = "promtail"
  namespace  = "monitoring"
  version    = "6.15.5"

  values = [
    yamlencode({
      config = {
        lokiAddress = "http://loki:3100/loki/api/v1/push"
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

resource "kubectl_manifest" "grafana_httproute" {
  yaml_body = yamlencode({
    apiVersion = "gateway.networking.k8s.io/v1"
    kind       = "HTTPRoute"
    metadata = {
      name      = "grafana-route"
      namespace = "monitoring"
    }
    spec = {
      parentRefs = [
        {
          name        = "nginx-gateway"
          namespace   = "nginx-gateway"
          sectionName = "https"
        }
      ]
      hostnames = ["grafana.${var.domain_name}"]
      rules = [
        {
          matches = [
            {
              path = {
                type  = "PathPrefix"
                value = "/"
              }
            }
          ]
          backendRefs = [
            {
              name = "prometheus-grafana"
              port = 80
            }
          ]
        }
      ]
    }
  })

  depends_on = [helm_release.kube_prometheus_stack]
}

resource "kubectl_manifest" "prometheus_httproute" {
  yaml_body = yamlencode({
    apiVersion = "gateway.networking.k8s.io/v1"
    kind       = "HTTPRoute"
    metadata = {
      name      = "prometheus-route"
      namespace = "monitoring"
    }
    spec = {
      parentRefs = [
        {
          name        = "nginx-gateway"
          namespace   = "nginx-gateway"
          sectionName = "https"
        }
      ]
      hostnames = ["prometheus.${var.domain_name}"]
      rules = [
        {
          matches = [
            {
              path = {
                type  = "PathPrefix"
                value = "/"
              }
            }
          ]
          backendRefs = [
            {
              name = "prometheus-kube-prometheus-prometheus"
              port = 9090
            }
          ]
        }
      ]
    }
  })

  depends_on = [helm_release.kube_prometheus_stack]
}

resource "kubectl_manifest" "loki_httproute" {
  yaml_body = yamlencode({
    apiVersion = "gateway.networking.k8s.io/v1"
    kind       = "HTTPRoute"
    metadata = {
      name      = "loki-route"
      namespace = "monitoring"
    }
    spec = {
      parentRefs = [
        {
          name        = "nginx-gateway"
          namespace   = "nginx-gateway"
          sectionName = "https"
        }
      ]
      hostnames = ["loki.${var.domain_name}"]
      rules = [
        {
          matches = [
            {
              path = {
                type  = "PathPrefix"
                value = "/"
              }
            }
          ]
          backendRefs = [
            {
              name = "loki"
              port = 3100
            }
          ]
        }
      ]
    }
  })

  depends_on = [helm_release.loki]
}

resource "kubectl_manifest" "rabbitmq_httproute" {
  yaml_body = yamlencode({
    apiVersion = "gateway.networking.k8s.io/v1"
    kind       = "HTTPRoute"
    metadata = {
      name      = "rabbitmq-management-route"
      namespace = data.terraform_remote_state.kubernetes_base.outputs.app_namespace
    }
    spec = {
      parentRefs = [
        {
          name        = "nginx-gateway"
          namespace   = "nginx-gateway"
          sectionName = "https"
        }
      ]
      hostnames = ["rabbitmq.${var.domain_name}"]
      rules = [
        {
          matches = [
            {
              path = {
                type  = "PathPrefix"
                value = "/"
              }
            }
          ]
          backendRefs = [
            {
              name = "rabbitmq"
              port = 15672
            }
          ]
        }
      ]
    }
  })
}

