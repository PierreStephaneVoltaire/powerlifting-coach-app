resource "kubernetes_namespace" "secrets" {
  metadata {
    name = "secrets"
  }
}
resource "kubernetes_namespace" "data" {
  metadata {
    name = "data"
  }
}


# resource "helm_release" "nginx_ingress" {
#   name             = "nginx-ingress"
#   repository       = "https://kubernetes.github.io/ingress-nginx"
#   chart            = "ingress-nginx"
#   namespace        = "ingress-nginx"
#   create_namespace = true
#   wait             = true
#   wait_for_jobs    = true
#   set {
#     name  = "controller.service.type"
#     value = "LoadBalancer"
#   }

#   set {
#     name  = "controller.service.annotations.service\\.beta\\.kubernetes\\.io/do-loadbalancer-name"
#     value = "${local.cluster_name}-lb"
#   }

#   set {
#     name  = "controller.service.annotations.service\\.beta\\.kubernetes\\.io/do-loadbalancer-protocol"
#     value = "http"
#   }

#   # Enable HTTP/2
#   set {
#     name  = "controller.service.annotations.service\\.beta\\.kubernetes\\.io/do-loadbalancer-enable-proxy-protocol"
#     value = "true"
#   }
#   timeout = 25 * 60
#   depends_on = [
#     digitalocean_kubernetes_cluster.k8s
#   ]
# }

# data "kubernetes_service" "nginx_ingress" {
#   metadata {
#     name      = "nginx-ingress-ingress-nginx-controller"
#     namespace = "ingress-nginx"
#   }

#   depends_on = [
#     helm_release.nginx_ingress
#   ]
# }

# output "load_balancer_ip" {
#   value       = data.kubernetes_service.nginx_ingress.status[0].load_balancer[0].ingress[0].ip
#   description = "Load balancer IP address"
# }