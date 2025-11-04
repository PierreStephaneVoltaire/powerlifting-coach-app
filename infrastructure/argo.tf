resource "kubernetes_namespace" "argocd" {
  metadata {
    name = "argocd"
  }
}
resource "kubernetes_namespace" "app" {
  metadata {
    name = "app"
  }
}


resource "helm_release" "argocd" {
  name       = "argocd"
  repository = "https://argoproj.github.io/argo-helm"
  chart      = "argo-cd"
  version    = "v9.0.5"
  namespace  = kubernetes_namespace.argocd.metadata[0].name
  depends_on = [
    kubernetes_namespace.argocd
  ]
}
resource "kubernetes_manifest" "app" {
  manifest = {
    "apiVersion" = "argoproj.io/v1alpha1"
    "kind"       = "Application"
    "metadata" = {
      "name"      = "${var.project_name}"
      "namespace" = kubernetes_namespace.argocd.metadata[0].name
    }
    "spec" = {
      "project" = "default"

      "source" = {
        "repoURL"        = "https://github.com/PierreStephaneVoltaire/powerlifting-coach-app"
        "path"           = "./k8s/overlays/production"
        "targetRevision" = "HEAD"
      }

      "destination" = {
        "server"    = "https://kubernetes.default.svc"
        "namespace" = "app"
      }


      "syncPolicy" = {
        "automated" = {
          "prune"    = true
          "selfHeal" = true
        }
        "syncOptions" = [
          "CreateNamespace=true"
        ]
      }
    }
  }
  depends_on = [
    helm_release.argocd
  ]
}