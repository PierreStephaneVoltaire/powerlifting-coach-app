resource "kubernetes_service_account" "argocd_app" {
  metadata {
    name      = "argocd-application-controller"
    namespace = data.terraform_remote_state.kubernetes_base.outputs.app_namespace
  }
}

resource "kubernetes_role" "argocd_app" {
  metadata {
    name      = "argocd-application-controller"
    namespace = data.terraform_remote_state.kubernetes_base.outputs.app_namespace
  }

  rule {
    api_groups = ["*"]
    resources  = ["*"]
    verbs      = ["*"]
  }
}

resource "kubernetes_role_binding" "argocd_app" {
  metadata {
    name      = "argocd-application-controller"
    namespace = data.terraform_remote_state.kubernetes_base.outputs.app_namespace
  }

  role_ref {
    api_group = "rbac.authorization.k8s.io"
    kind      = "Role"
    name      = kubernetes_role.argocd_app.metadata[0].name
  }

  subject {
    kind      = "ServiceAccount"
    name      = "argocd-application-controller"
    namespace = data.terraform_remote_state.argocd.outputs.argocd_namespace
  }
}

resource "kubernetes_cluster_role" "argocd_server" {
  metadata {
    name = "argocd-server-cluster-apps"
  }

  rule {
    api_groups = [""]
    resources  = ["namespaces"]
    verbs      = ["get", "list", "watch"]
  }

  rule {
    api_groups = ["argoproj.io"]
    resources  = ["applications", "appprojects"]
    verbs      = ["*"]
  }
}

resource "kubernetes_cluster_role_binding" "argocd_server" {
  metadata {
    name = "argocd-server-cluster-apps"
  }

  role_ref {
    api_group = "rbac.authorization.k8s.io"
    kind      = "ClusterRole"
    name      = kubernetes_cluster_role.argocd_server.metadata[0].name
  }

  subject {
    kind      = "ServiceAccount"
    name      = "argocd-server"
    namespace = data.terraform_remote_state.argocd.outputs.argocd_namespace
  }
}

resource "kubernetes_manifest" "app_frontend" {
  count = var.deploy_frontend ? 1 : 0

  field_manager {
    force_conflicts = true
  }

  manifest = {
    "apiVersion" = "argoproj.io/v1alpha1"
    "kind"       = "Application"
    "metadata" = {
      "name"      = "${var.project_name}-frontend"
      "namespace" = data.terraform_remote_state.argocd.outputs.argocd_namespace
      "annotations" = {
        "argocd-image-updater.argoproj.io/image-list"               = "frontend=ghcr.io/pierrestephanevoltaire/powerlifting-coach/frontend"
        "argocd-image-updater.argoproj.io/write-back-method"        = "argocd"
        "argocd-image-updater.argoproj.io/write-back-target"        = "kustomization"
        "argocd-image-updater.argoproj.io/frontend.update-strategy" = "digest"
        "argocd-image-updater.argoproj.io/frontend.allow-tags"      = "regexp:^.*$"
        "argocd-image-updater.argoproj.io/frontend.force-update"    = "true"
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
}

resource "kubernetes_manifest" "app_datalayer" {
  count = var.deploy_datalayer ? 1 : 0

  field_manager {
    force_conflicts = true
  }

  manifest = {
    "apiVersion" = "argoproj.io/v1alpha1"
    "kind"       = "Application"
    "metadata" = {
      "name"      = "${var.project_name}-datalayer"
      "namespace" = data.terraform_remote_state.argocd.outputs.argocd_namespace
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
}

resource "kubernetes_manifest" "app_backend" {
  count = var.deploy_backend ? 1 : 0

  field_manager {
    force_conflicts = true
  }

  manifest = {
    "apiVersion" = "argoproj.io/v1alpha1"
    "kind"       = "Application"
    "metadata" = {
      "name"      = "${var.project_name}-backend"
      "namespace" = data.terraform_remote_state.argocd.outputs.argocd_namespace
      "annotations" = {
        "argocd-image-updater.argoproj.io/image-list" = join(",", [
          "auth-service=ghcr.io/pierrestephanevoltaire/powerlifting-coach/auth-service",
          "user-service=ghcr.io/pierrestephanevoltaire/powerlifting-coach/user-service",
          "video-service=ghcr.io/pierrestephanevoltaire/powerlifting-coach/video-service",
          "media-processor-service=ghcr.io/pierrestephanevoltaire/powerlifting-coach/media-processor-service",
          "settings-service=ghcr.io/pierrestephanevoltaire/powerlifting-coach/settings-service",
          "program-service=ghcr.io/pierrestephanevoltaire/powerlifting-coach/program-service",
          "coach-service=ghcr.io/pierrestephanevoltaire/powerlifting-coach/coach-service",
          "notification-service=ghcr.io/pierrestephanevoltaire/powerlifting-coach/notification-service",
          "dm-service=ghcr.io/pierrestephanevoltaire/powerlifting-coach/dm-service",
          "machine-service=ghcr.io/pierrestephanevoltaire/powerlifting-coach/machine-service",
          "reminder-service=ghcr.io/pierrestephanevoltaire/powerlifting-coach/reminder-service"
        ])
        "argocd-image-updater.argoproj.io/write-back-method"                       = "argocd"
        "argocd-image-updater.argoproj.io/write-back-target"                      = "kustomization"
        "argocd-image-updater.argoproj.io/auth-service.update-strategy"            = "digest"
        "argocd-image-updater.argoproj.io/user-service.update-strategy"            = "digest"
        "argocd-image-updater.argoproj.io/video-service.update-strategy"           = "digest"
        "argocd-image-updater.argoproj.io/media-processor-service.update-strategy" = "digest"
        "argocd-image-updater.argoproj.io/settings-service.update-strategy"        = "digest"
        "argocd-image-updater.argoproj.io/program-service.update-strategy"         = "digest"
        "argocd-image-updater.argoproj.io/coach-service.update-strategy"           = "digest"
        "argocd-image-updater.argoproj.io/notification-service.update-strategy"    = "digest"
        "argocd-image-updater.argoproj.io/dm-service.update-strategy"              = "digest"
        "argocd-image-updater.argoproj.io/machine-service.update-strategy"         = "digest"
        "argocd-image-updater.argoproj.io/reminder-service.update-strategy"        = "digest"
        "argocd-image-updater.argoproj.io/auth-service.allow-tags"                 = "regexp:^.*$"
        "argocd-image-updater.argoproj.io/user-service.allow-tags"                 = "regexp:^.*$"
        "argocd-image-updater.argoproj.io/video-service.allow-tags"                = "regexp:^.*$"
        "argocd-image-updater.argoproj.io/media-processor-service.allow-tags"      = "regexp:^.*$"
        "argocd-image-updater.argoproj.io/settings-service.allow-tags"             = "regexp:^.*$"
        "argocd-image-updater.argoproj.io/program-service.allow-tags"              = "regexp:^.*$"
        "argocd-image-updater.argoproj.io/coach-service.allow-tags"                = "regexp:^.*$"
        "argocd-image-updater.argoproj.io/notification-service.allow-tags"         = "regexp:^.*$"
        "argocd-image-updater.argoproj.io/dm-service.allow-tags"                   = "regexp:^.*$"
        "argocd-image-updater.argoproj.io/machine-service.allow-tags"              = "regexp:^.*$"
        "argocd-image-updater.argoproj.io/reminder-service.allow-tags"             = "regexp:^.*$"
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
}
