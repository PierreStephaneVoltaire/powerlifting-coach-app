resource "helm_release" "nginx_ingress" {
  count = var.kubernetes_resources_enabled ? 1 : 0

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
    name  = "controller.service.annotations.service\\.beta\\.kubernetes\\.io/do-loadbalancer-name"
    value = "${local.cluster_name}-lb"
  }

  set {
    name  = "controller.service.annotations.service\\.beta\\.kubernetes\\.io/do-loadbalancer-protocol"
    value = "http"
  }

  set {
    name  = "controller.service.annotations.service\\.beta\\.kubernetes\\.io/do-loadbalancer-enable-proxy-protocol"
    value = "true"
  }
  timeout = 25 * 60
  depends_on = [
    digitalocean_kubernetes_cluster.k8s
  ]
}

data "kubernetes_service" "nginx_ingress" {
  count = var.kubernetes_resources_enabled ? 1 : 0

  metadata {
    name      = "nginx-ingress-ingress-nginx-controller"
    namespace = "ingress-nginx"
  }

  depends_on = [
    helm_release.nginx_ingress
  ]
}

output "load_balancer_ip" {
  value       = var.kubernetes_resources_enabled ? data.kubernetes_service.nginx_ingress[0].status[0].load_balancer[0].ingress[0].ip : "not-yet-available"
  description = "Load balancer IP address"
}

resource "helm_release" "metrics_server" {
  count = var.kubernetes_resources_enabled ? 1 : 0

  name             = "metrics-server"
  repository       = "https://kubernetes-sigs.github.io/metrics-server/"
  chart            = "metrics-server"
  namespace        = "kube-system"
  create_namespace = false
  wait             = true
  wait_for_jobs    = true

  set {
    name  = "args[0]"
    value = "--kubelet-insecure-tls"
  }

  set {
    name  = "args[1]"
    value = "--kubelet-preferred-address-types=InternalIP"
  }

  timeout = 10 * 60
  depends_on = [
    digitalocean_kubernetes_cluster.k8s
  ]
}