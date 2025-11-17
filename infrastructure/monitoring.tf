resource "kubectl_manifest" "grafana_httproute" {
  count = var.kubernetes_resources_enabled ? 1 : 0

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
          name      = "nginx-gateway"
          namespace = "nginx-gateway"
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

  depends_on = [
    helm_release.kube_prometheus_stack,
    helm_release.nginx_gateway_fabric
  ]
}

resource "kubectl_manifest" "grafana_certificate" {
  count = var.kubernetes_resources_enabled ? 1 : 0

  yaml_body = yamlencode({
    apiVersion = "cert-manager.io/v1"
    kind       = "Certificate"
    metadata = {
      name      = "grafana-tls"
      namespace = "nginx-gateway"
    }
    spec = {
      secretName = "grafana-tls"
      issuerRef = {
        name = "letsencrypt-prod"
        kind = "ClusterIssuer"
      }
      dnsNames = ["grafana.${var.domain_name}"]
    }
  })

  depends_on = [kubectl_manifest.letsencrypt_prod]
}

resource "kubectl_manifest" "prometheus_httproute" {
  count = var.kubernetes_resources_enabled ? 1 : 0

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
          name      = "nginx-gateway"
          namespace = "nginx-gateway"
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

  depends_on = [
    helm_release.kube_prometheus_stack,
    helm_release.nginx_gateway_fabric
  ]
}

resource "kubectl_manifest" "prometheus_certificate" {
  count = var.kubernetes_resources_enabled ? 1 : 0

  yaml_body = yamlencode({
    apiVersion = "cert-manager.io/v1"
    kind       = "Certificate"
    metadata = {
      name      = "prometheus-tls"
      namespace = "nginx-gateway"
    }
    spec = {
      secretName = "prometheus-tls"
      issuerRef = {
        name = "letsencrypt-prod"
        kind = "ClusterIssuer"
      }
      dnsNames = ["prometheus.${var.domain_name}"]
    }
  })

  depends_on = [kubectl_manifest.letsencrypt_prod]
}

resource "kubectl_manifest" "loki_httproute" {
  count = var.kubernetes_resources_enabled ? 1 : 0

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
          name      = "nginx-gateway"
          namespace = "nginx-gateway"
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

  depends_on = [
    helm_release.loki,
    helm_release.nginx_gateway_fabric
  ]
}

resource "kubectl_manifest" "loki_certificate" {
  count = var.kubernetes_resources_enabled ? 1 : 0

  yaml_body = yamlencode({
    apiVersion = "cert-manager.io/v1"
    kind       = "Certificate"
    metadata = {
      name      = "loki-tls"
      namespace = "nginx-gateway"
    }
    spec = {
      secretName = "loki-tls"
      issuerRef = {
        name = "letsencrypt-prod"
        kind = "ClusterIssuer"
      }
      dnsNames = ["loki.${var.domain_name}"]
    }
  })

  depends_on = [kubectl_manifest.letsencrypt_prod]
}

resource "kubectl_manifest" "rabbitmq_httproute" {
  count = var.kubernetes_resources_enabled ? 1 : 0

  yaml_body = yamlencode({
    apiVersion = "gateway.networking.k8s.io/v1"
    kind       = "HTTPRoute"
    metadata = {
      name      = "rabbitmq-management-route"
      namespace = kubernetes_namespace.app[0].metadata[0].name
    }
    spec = {
      parentRefs = [
        {
          name      = "nginx-gateway"
          namespace = "nginx-gateway"
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

  depends_on = [
    kubernetes_namespace.app,
    helm_release.nginx_gateway_fabric
  ]
}

resource "kubectl_manifest" "rabbitmq_certificate" {
  count = var.kubernetes_resources_enabled ? 1 : 0

  yaml_body = yamlencode({
    apiVersion = "cert-manager.io/v1"
    kind       = "Certificate"
    metadata = {
      name      = "rabbitmq-tls"
      namespace = "nginx-gateway"
    }
    spec = {
      secretName = "rabbitmq-tls"
      issuerRef = {
        name = "letsencrypt-prod"
        kind = "ClusterIssuer"
      }
      dnsNames = ["rabbitmq.${var.domain_name}"]
    }
  })

  depends_on = [kubectl_manifest.letsencrypt_prod]
}
