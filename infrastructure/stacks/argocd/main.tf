resource "kubernetes_namespace" "argocd" {
  metadata {
    name = "argocd"
  }
}

data "kubernetes_secret" "argocd_admin" {
  count = var.stopped ? 0 : 1

  metadata {
    name      = "argocd-initial-admin-secret"
    namespace = kubernetes_namespace.argocd.metadata[0].name
  }

  depends_on = [helm_release.argocd]
}

resource "aws_ssm_parameter" "argocd_admin_password" {
  count = var.stopped ? 0 : 1

  name  = "/${var.project_name}/${var.environment}/argocd/admin-password"
  type  = "SecureString"
  value = data.kubernetes_secret.argocd_admin[0].data["password"]

  depends_on = [data.kubernetes_secret.argocd_admin]
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

