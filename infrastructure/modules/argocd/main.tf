resource "kubernetes_namespace" "argocd" {
  metadata {
    name = "argocd"
  }
}

resource "helm_release" "argocd" {
  count = var.stopped ? 0 : 1

  name          = "argocd"
  repository    = "https://argoproj.github.io/argo-helm"
  chart         = "argo-cd"
  version       = "v9.0.5"
  namespace     = kubernetes_namespace.argocd.metadata[0].name
  wait          = true
  wait_for_jobs = true
  timeout       = 600

  set {
    name  = "configs.params.server\\.insecure"
    value = "true"
  }

  set {
    name  = "configs.cm.kustomize\\.buildOptions"
    value = "--load-restrictor LoadRestrictionsNone"
  }

  set {
    name  = "global.domain"
    value = "argocd.${var.domain_name}"
  }

  depends_on = [
    kubernetes_namespace.argocd
  ]
}

resource "kubectl_manifest" "argocd_httproute" {
  yaml_body = yamlencode({
    apiVersion = "gateway.networking.k8s.io/v1"
    kind       = "HTTPRoute"
    metadata = {
      name      = "argocd-route"
      namespace = kubernetes_namespace.argocd.metadata[0].name
    }
    spec = {
      parentRefs = [
        {
          name        = "nginx-gateway"
          namespace   = "nginx-gateway"
          sectionName = "https"
        }
      ]
      hostnames = ["argocd.${var.domain_name}"]
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
              name = "argocd-server"
              port = 80
            }
          ]
        }
      ]
    }
  })

  depends_on = [
    helm_release.argocd
  ]
}

