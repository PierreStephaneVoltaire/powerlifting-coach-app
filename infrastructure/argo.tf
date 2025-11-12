resource "kubernetes_namespace" "argocd" {
  count = var.kubernetes_resources_enabled ? 1 : 0

  metadata {
    name = "argocd"
  }
}

resource "helm_release" "argocd" {
  count = var.kubernetes_resources_enabled && !var.stopped ? 1 : 0

  name          = "argocd"
  repository    = "https://argoproj.github.io/argo-helm"
  chart         = "argo-cd"
  version       = "v9.0.5"
  namespace     = kubernetes_namespace.argocd[0].metadata[0].name
  wait          = true
  wait_for_jobs = true
  timeout       = 600

  set {
    name  = "configs.params.server\\.insecure"
    value = "true"
  }

  depends_on = [
    kubernetes_namespace.argocd
  ]
}

resource "kubernetes_ingress_v1" "argocd" {
  count = var.kubernetes_resources_enabled ? 1 : 0

  metadata {
    name      = "argocd-ingress"
    namespace = kubernetes_namespace.argocd[0].metadata[0].name
    annotations = {
      "kubernetes.io/ingress.class"                    = "nginx"
      "nginx.ingress.kubernetes.io/backend-protocol"   = "HTTP"
      "cert-manager.io/cluster-issuer"                 = "letsencrypt-prod"
      "nginx.ingress.kubernetes.io/ssl-redirect"       = "true"
      "nginx.ingress.kubernetes.io/force-ssl-redirect" = "true"
    }
  }

  spec {
    tls {
      hosts       = ["argocd.${var.domain_name}"]
      secret_name = "argocd-tls"
    }

    rule {
      host = "argocd.${var.domain_name}"
      http {
        path {
          path      = "/"
          path_type = "Prefix"
          backend {
            service {
              name = "argocd-server"
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
    helm_release.argocd,
    data.kubernetes_service.nginx_ingress
  ]
}

resource "kubernetes_manifest" "app" {
  count = var.kubernetes_resources_enabled && var.argocd_resources_enabled ? 1 : 0

  field_manager {
    force_conflicts = true
  }

  manifest = {
    "apiVersion" = "argoproj.io/v1alpha1"
    "kind"       = "Application"
    "metadata" = {
      "name"      = "${var.project_name}"
      "namespace" = kubernetes_namespace.argocd[0].metadata[0].name
      "annotations" = {
        # Configure ArgoCD Image Updater with Kubernetes API write-back
        "argocd-image-updater.argoproj.io/image-list" = join(",", [
          "auth-service=ghcr.io/pierrestephanevoltaire/powerlifting-coach/auth-service:latest",
          "user-service=ghcr.io/pierrestephanevoltaire/powerlifting-coach/user-service:latest",
          "video-service=ghcr.io/pierrestephanevoltaire/powerlifting-coach/video-service:latest",
          "settings-service=ghcr.io/pierrestephanevoltaire/powerlifting-coach/settings-service:latest",
          "program-service=ghcr.io/pierrestephanevoltaire/powerlifting-coach/program-service:latest",
          "coach-service=ghcr.io/pierrestephanevoltaire/powerlifting-coach/coach-service:latest",
          "notification-service=ghcr.io/pierrestephanevoltaire/powerlifting-coach/notification-service:latest",
          "frontend=ghcr.io/pierrestephanevoltaire/powerlifting-coach/frontend:latest"
        ])
        "argocd-image-updater.argoproj.io/write-back-method" = "argocd"
        "argocd-image-updater.argoproj.io/auth-service.update-strategy" = "latest"
        "argocd-image-updater.argoproj.io/user-service.update-strategy" = "latest"
        "argocd-image-updater.argoproj.io/video-service.update-strategy" = "latest"
        "argocd-image-updater.argoproj.io/settings-service.update-strategy" = "latest"
        "argocd-image-updater.argoproj.io/program-service.update-strategy" = "latest"
        "argocd-image-updater.argoproj.io/coach-service.update-strategy" = "latest"
        "argocd-image-updater.argoproj.io/notification-service.update-strategy" = "latest"
        "argocd-image-updater.argoproj.io/frontend.update-strategy" = "latest"
      }
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