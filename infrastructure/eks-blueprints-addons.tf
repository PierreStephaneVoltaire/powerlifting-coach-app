data "aws_ecrpublic_authorization_token" "token" {
  provider = aws.virginia
}



resource "helm_release" "cert_manager" {
  count = var.kubernetes_resources_enabled ? 1 : 0

  name             = "cert-manager"
  repository       = "https://charts.jetstack.io"
  chart            = "cert-manager"
  namespace        = "cert-manager"
  create_namespace = true
  version          = "v1.14.2"

  values = [
    yamlencode({
      installCRDs = true
      resources = {
        requests = {
          cpu    = "50m"
          memory = "128Mi"
        }
        limits = {
          cpu    = "100m"
          memory = "256Mi"
        }
      }
    })
  ]
}

resource "helm_release" "external_dns" {
  count = var.kubernetes_resources_enabled ? 1 : 0

  name             = "external-dns"
  repository       = "https://kubernetes-sigs.github.io/external-dns/"
  chart            = "external-dns"
  namespace        = "kube-system"
  version          = "1.14.3"

  values = [
    yamlencode({
      provider      = "aws"
      domainFilters = [var.domain_name]
      policy        = "sync"
      registry      = "txt"
      txtOwnerId    = local.cluster_name
      interval      = "1m"
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
      serviceAccount = {
        create = true
        annotations = {
        }
      }
    })
  ]

  depends_on = [helm_release.cert_manager]
}

resource "helm_release" "kube_prometheus_stack" {
  count = var.kubernetes_resources_enabled && !var.stopped ? 1 : 0

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
        adminPassword = var.kubernetes_resources_enabled ? random_password.grafana_admin_password[0].result : ""
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

  depends_on = [helm_release.cert_manager]
}

resource "kubectl_manifest" "letsencrypt_prod" {
  count = var.kubernetes_resources_enabled ? 1 : 0

  yaml_body = yamlencode({
    apiVersion = "cert-manager.io/v1"
    kind       = "ClusterIssuer"
    metadata = {
      name = "letsencrypt-prod"
    }
    spec = {
      acme = {
        server = "https://acme-v02.api.letsencrypt.org/directory"
        email  = "admin@${var.domain_name}"
        privateKeySecretRef = {
          name = "letsencrypt-prod"
        }
        solvers = [
          {
            http01 = {
              gatewayHTTPRoute = {
                parentRefs = [
                  {
                    name      = "nginx-gateway"
                    namespace = "nginx-gateway"
                    kind      = "Gateway"
                  }
                ]
              }
            }
          }
        ]
      }
    }
  })

  depends_on = [helm_release.cert_manager, helm_release.nginx_gateway_fabric]
}

resource "helm_release" "gateway_api_crds" {
  count = var.kubernetes_resources_enabled && !var.stopped ? 1 : 0

  name             = "gateway-api"
  repository       = "https://kubernetes-sigs.github.io/gateway-api/charts"
  chart            = "gateway-api"
  namespace        = "gateway-system"
  create_namespace = true
  version          = "1.2.0"
}

resource "helm_release" "nginx_gateway_fabric" {
  count = var.kubernetes_resources_enabled && !var.stopped ? 1 : 0

  name             = "nginx-gateway-fabric"
  repository       = "oci://ghcr.io/nginxinc/charts"
  chart            = "nginx-gateway-fabric"
  namespace        = "nginx-gateway"
  create_namespace = true
  version          = "1.4.0"

  values = [
    yamlencode({
      nginxGateway = {
        gatewayClassName = "nginx"
        config = {
          logging = {
            level = "debug"
          }
        }
      }
      nginx = {
        config = {
          entries = [
            {
              name  = "proxy_intercept_errors"
              value = "off"
            },
            {
              name  = "proxy_connect_timeout"
              value = "60s"
            },
            {
              name  = "proxy_send_timeout"
              value = "60s"
            },
            {
              name  = "proxy_read_timeout"
              value = "60s"
            },
            {
              name  = "client_max_body_size"
              value = "100m"
            }
          ]
        }
        resources = {
          requests = {
            cpu    = "100m"
            memory = "128Mi"
          }
          limits = {
            cpu    = "200m"
            memory = "256Mi"
          }
        }
      }
      service = {
        type = "LoadBalancer"
      }
    })
  ]

  depends_on = [helm_release.gateway_api_crds]
}

resource "kubectl_manifest" "nginx_gateway" {
  count = var.kubernetes_resources_enabled && !var.stopped ? 1 : 0

  yaml_body = yamlencode({
    apiVersion = "gateway.networking.k8s.io/v1"
    kind       = "Gateway"
    metadata = {
      name      = "nginx-gateway"
      namespace = "nginx-gateway"
      annotations = {
        "cert-manager.io/cluster-issuer" = "letsencrypt-prod"
      }
    }
    spec = {
      gatewayClassName = "nginx"
      listeners = [
        {
          name     = "http"
          port     = 80
          protocol = "HTTP"
          allowedRoutes = {
            namespaces = {
              from = "All"
            }
          }
        },
        {
          name     = "https"
          port     = 443
          protocol = "HTTPS"
          tls = {
            mode = "Terminate"
            certificateRefs = [
              {
                kind = "Secret"
                name = "wildcard-tls"
              }
            ]
          }
          allowedRoutes = {
            namespaces = {
              from = "All"
            }
          }
        }
      ]
    }
  })

  depends_on = [helm_release.nginx_gateway_fabric]
}

resource "helm_release" "loki" {
  count = var.kubernetes_resources_enabled && !var.stopped ? 1 : 0

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
  count = var.kubernetes_resources_enabled && !var.stopped ? 1 : 0

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
