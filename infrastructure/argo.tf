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
  count = var.kubernetes_resources_enabled ? 1 : 0

  yaml_body = yamlencode({
    apiVersion = "gateway.networking.k8s.io/v1"
    kind       = "HTTPRoute"
    metadata = {
      name      = "argocd-route"
      namespace = kubernetes_namespace.argocd[0].metadata[0].name
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
    helm_release.argocd,
    helm_release.nginx_gateway_fabric
  ]
}

resource "kubectl_manifest" "argocd_certificate" {
  count = var.kubernetes_resources_enabled ? 1 : 0

  yaml_body = yamlencode({
    apiVersion = "cert-manager.io/v1"
    kind       = "Certificate"
    metadata = {
      name      = "argocd-tls"
      namespace = "nginx-gateway"
    }
    spec = {
      secretName = "argocd-tls"
      issuerRef = {
        name = "letsencrypt-prod"
        kind = "ClusterIssuer"
      }
      dnsNames = ["argocd.${var.domain_name}"]
    }
  })

  depends_on = [kubectl_manifest.letsencrypt_prod]
}

resource "kubernetes_manifest" "app_frontend" {
  count = var.kubernetes_resources_enabled && var.argocd_resources_enabled && var.deploy_frontend ? 1 : 0

  field_manager {
    force_conflicts = true
  }

  manifest = {
    "apiVersion" = "argoproj.io/v1alpha1"
    "kind"       = "Application"
    "metadata" = {
      "name"      = "${var.project_name}-frontend"
      "namespace" = kubernetes_namespace.argocd[0].metadata[0].name
      "annotations" = {
        "argocd-image-updater.argoproj.io/image-list" = "frontend=ghcr.io/pierrestephanevoltaire/powerlifting-coach/frontend:latest"
        "argocd-image-updater.argoproj.io/write-back-method" = "argocd"
        "argocd-image-updater.argoproj.io/frontend.update-strategy" = "latest"
      }
    }
    "spec" = {
      "project" = "default"

      "source" = {
        "repoURL"        = "https://github.com/PierreStephaneVoltaire/powerlifting-coach-app"
        "path"           = "./k8s/overlays/production-frontend"
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

resource "kubernetes_manifest" "app_datalayer" {
  count = var.kubernetes_resources_enabled && var.argocd_resources_enabled && var.deploy_datalayer ? 1 : 0

  field_manager {
    force_conflicts = true
  }

  manifest = {
    "apiVersion" = "argoproj.io/v1alpha1"
    "kind"       = "Application"
    "metadata" = {
      "name"      = "${var.project_name}-datalayer"
      "namespace" = kubernetes_namespace.argocd[0].metadata[0].name
    }
    "spec" = {
      "project" = "default"

      "source" = {
        "repoURL"        = "https://github.com/PierreStephaneVoltaire/powerlifting-coach-app"
        "path"           = "./k8s/overlays/production-datalayer"
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

resource "kubernetes_manifest" "app_backend" {
  count = var.kubernetes_resources_enabled && var.argocd_resources_enabled && var.deploy_backend ? 1 : 0

  field_manager {
    force_conflicts = true
  }

  manifest = {
    "apiVersion" = "argoproj.io/v1alpha1"
    "kind"       = "Application"
    "metadata" = {
      "name"      = "${var.project_name}-backend"
      "namespace" = kubernetes_namespace.argocd[0].metadata[0].name
      "annotations" = {
        "argocd-image-updater.argoproj.io/image-list" = join(",", [
          "auth-service=ghcr.io/pierrestephanevoltaire/powerlifting-coach/auth-service:latest",
          "user-service=ghcr.io/pierrestephanevoltaire/powerlifting-coach/user-service:latest",
          "video-service=ghcr.io/pierrestephanevoltaire/powerlifting-coach/video-service:latest",
          "media-processor-service=ghcr.io/pierrestephanevoltaire/powerlifting-coach/media-processor-service:latest",
          "settings-service=ghcr.io/pierrestephanevoltaire/powerlifting-coach/settings-service:latest",
          "program-service=ghcr.io/pierrestephanevoltaire/powerlifting-coach/program-service:latest",
          "coach-service=ghcr.io/pierrestephanevoltaire/powerlifting-coach/coach-service:latest",
          "notification-service=ghcr.io/pierrestephanevoltaire/powerlifting-coach/notification-service:latest",
          "dm-service=ghcr.io/pierrestephanevoltaire/powerlifting-coach/dm-service:latest",
          "machine-service=ghcr.io/pierrestephanevoltaire/powerlifting-coach/machine-service:latest",
          "reminder-service=ghcr.io/pierrestephanevoltaire/powerlifting-coach/reminder-service:latest"
        ])
        "argocd-image-updater.argoproj.io/write-back-method" = "argocd"
        "argocd-image-updater.argoproj.io/auth-service.update-strategy" = "latest"
        "argocd-image-updater.argoproj.io/user-service.update-strategy" = "latest"
        "argocd-image-updater.argoproj.io/video-service.update-strategy" = "latest"
        "argocd-image-updater.argoproj.io/media-processor-service.update-strategy" = "latest"
        "argocd-image-updater.argoproj.io/settings-service.update-strategy" = "latest"
        "argocd-image-updater.argoproj.io/program-service.update-strategy" = "latest"
        "argocd-image-updater.argoproj.io/coach-service.update-strategy" = "latest"
        "argocd-image-updater.argoproj.io/notification-service.update-strategy" = "latest"
        "argocd-image-updater.argoproj.io/dm-service.update-strategy" = "latest"
        "argocd-image-updater.argoproj.io/machine-service.update-strategy" = "latest"
        "argocd-image-updater.argoproj.io/reminder-service.update-strategy" = "latest"
      }
    }
    "spec" = {
      "project" = "default"

      "source" = {
        "repoURL"        = "https://github.com/PierreStephaneVoltaire/powerlifting-coach-app"
        "path"           = "./k8s/overlays/production-backend"
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
