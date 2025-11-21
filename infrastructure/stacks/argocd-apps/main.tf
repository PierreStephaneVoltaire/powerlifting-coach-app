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
        "argocd-image-updater.argoproj.io/image-list"               = "frontend=ghcr.io/pierrestephanevoltaire/powerlifting-coach/frontend:latest"
        "argocd-image-updater.argoproj.io/write-back-method"        = "argocd"
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
        "argocd-image-updater.argoproj.io/write-back-method"                       = "argocd"
        "argocd-image-updater.argoproj.io/auth-service.update-strategy"            = "latest"
        "argocd-image-updater.argoproj.io/user-service.update-strategy"            = "latest"
        "argocd-image-updater.argoproj.io/video-service.update-strategy"           = "latest"
        "argocd-image-updater.argoproj.io/media-processor-service.update-strategy" = "latest"
        "argocd-image-updater.argoproj.io/settings-service.update-strategy"        = "latest"
        "argocd-image-updater.argoproj.io/program-service.update-strategy"         = "latest"
        "argocd-image-updater.argoproj.io/coach-service.update-strategy"           = "latest"
        "argocd-image-updater.argoproj.io/notification-service.update-strategy"    = "latest"
        "argocd-image-updater.argoproj.io/dm-service.update-strategy"              = "latest"
        "argocd-image-updater.argoproj.io/machine-service.update-strategy"         = "latest"
        "argocd-image-updater.argoproj.io/reminder-service.update-strategy"        = "latest"
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
