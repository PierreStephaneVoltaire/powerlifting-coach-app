resource "helm_release" "nginx_ingress" {
  count = var.kubernetes_resources_enabled && !var.stopped ? 1 : 0

  name             = "nginx-ingress"
  repository       = "https://kubernetes.github.io/ingress-nginx"
  chart            = "ingress-nginx"
  namespace        = "ingress-nginx"
  create_namespace = true
  wait             = true
  wait_for_jobs    = true
  set {
    name  = "controller.service.type"
    value = "LoadBalancer"
  }

  set {
    name  = "controller.service.annotations.service\\.beta\\.kubernetes\\.io/azure-load-balancer-health-probe-request-path"
    value = "/healthz"
  }
  timeout = 25 * 60
  depends_on = [
    azurerm_kubernetes_cluster.k8s
  ]
}

data "kubernetes_service" "nginx_ingress" {
  count = var.kubernetes_resources_enabled && !var.stopped ? 1 : 0

  metadata {
    name      = "nginx-ingress-ingress-nginx-controller"
    namespace = "ingress-nginx"
  }

  depends_on = [
    helm_release.nginx_ingress
  ]
}

output "load_balancer_ip" {
  value       = var.kubernetes_resources_enabled && !var.stopped ? data.kubernetes_service.nginx_ingress[0].status[0].load_balancer[0].ingress[0].ip : "stopped"
  description = "Load balancer IP address"
}

resource "helm_release" "argocd_image_updater" {
  count = var.kubernetes_resources_enabled ? 1 : 0

  name             = "argocd-image-updater"
  repository       = "https://argoproj.github.io/argo-helm"
  chart            = "argocd-image-updater"
  namespace        = "argocd"
  create_namespace = false
  wait             = true
  wait_for_jobs    = true

  set {
    name  = "config.argocd.plaintext"
    value = "true"
  }

  set {
    name  = "config.argocd.serverAddress"
    value = "http://argocd-server.argocd.svc.cluster.local"
  }

  set {
    name  = "config.argocd.insecure"
    value = "true"
  }

  set {
    name  = "config.applicationsAPIKind"
    value = "kubernetes"
  }

  set {
    name  = "config.registries[0].name"
    value = "GitHub Container Registry"
  }

  set {
    name  = "config.registries[0].prefix"
    value = "ghcr.io"
  }

  set {
    name  = "config.registries[0].api_url"
    value = "https://ghcr.io"
  }

  set {
    name  = "config.registries[0].ping"
    value = "true"
  }

  set {
    name  = "config.logLevel"
    value = "debug"
  }

  timeout = 10 * 60
  depends_on = [
    helm_release.argocd
  ]
}

resource "helm_release" "cert_manager" {
  count = var.kubernetes_resources_enabled && var.domain_name != "localhost" ? 1 : 0

  name             = "cert-manager"
  repository       = "https://charts.jetstack.io"
  chart            = "cert-manager"
  namespace        = "cert-manager"
  create_namespace = true
  version          = "v1.13.3"
  wait             = true
  wait_for_jobs    = true

  set {
    name  = "installCRDs"
    value = "true"
  }

  set {
    name  = "prometheus.enabled"
    value = "true"
  }

  # Configure cert-manager to run on spot nodes
  set {
    name  = "tolerations[0].key"
    value = "kubernetes.azure.com/scalesetpriority"
  }
  set {
    name  = "tolerations[0].operator"
    value = "Equal"
  }
  set {
    name  = "tolerations[0].value"
    value = "spot"
  }
  set {
    name  = "tolerations[0].effect"
    value = "NoSchedule"
  }

  set {
    name  = "nodeSelector.workload-type"
    value = "spot"
  }

  set {
    name  = "cainjector.tolerations[0].key"
    value = "kubernetes.azure.com/scalesetpriority"
  }
  set {
    name  = "cainjector.tolerations[0].operator"
    value = "Equal"
  }
  set {
    name  = "cainjector.tolerations[0].value"
    value = "spot"
  }
  set {
    name  = "cainjector.tolerations[0].effect"
    value = "NoSchedule"
  }

  set {
    name  = "cainjector.nodeSelector.workload-type"
    value = "spot"
  }

  set {
    name  = "webhook.tolerations[0].key"
    value = "kubernetes.azure.com/scalesetpriority"
  }
  set {
    name  = "webhook.tolerations[0].operator"
    value = "Equal"
  }
  set {
    name  = "webhook.tolerations[0].value"
    value = "spot"
  }
  set {
    name  = "webhook.tolerations[0].effect"
    value = "NoSchedule"
  }

  set {
    name  = "webhook.nodeSelector.workload-type"
    value = "spot"
  }

  timeout = 10 * 60
  depends_on = [
    azurerm_kubernetes_cluster.k8s
  ]
}