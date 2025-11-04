resource "kubernetes_service_account" "argocd_app" {
  count = var.kubernetes_resources_enabled && var.argocd_resources_enabled ? 1 : 0

  metadata {
    name      = "argocd-application-controller"
    namespace = kubernetes_namespace.app[0].metadata[0].name
  }
}

resource "kubernetes_role" "argocd_app" {
  count = var.kubernetes_resources_enabled && var.argocd_resources_enabled ? 1 : 0

  metadata {
    name      = "argocd-application-controller"
    namespace = kubernetes_namespace.app[0].metadata[0].name
  }

  rule {
    api_groups = ["*"]
    resources  = ["*"]
    verbs      = ["*"]
  }
}

resource "kubernetes_role_binding" "argocd_app" {
  count = var.kubernetes_resources_enabled && var.argocd_resources_enabled ? 1 : 0

  metadata {
    name      = "argocd-application-controller"
    namespace = kubernetes_namespace.app[0].metadata[0].name
  }

  role_ref {
    api_group = "rbac.authorization.k8s.io"
    kind      = "Role"
    name      = kubernetes_role.argocd_app[0].metadata[0].name
  }

  subject {
    kind      = "ServiceAccount"
    name      = "argocd-application-controller"
    namespace = kubernetes_namespace.argocd[0].metadata[0].name
  }
}

resource "kubernetes_cluster_role" "argocd_server" {
  count = var.kubernetes_resources_enabled && var.argocd_resources_enabled ? 1 : 0

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
  count = var.kubernetes_resources_enabled && var.argocd_resources_enabled ? 1 : 0

  metadata {
    name = "argocd-server-cluster-apps"
  }

  role_ref {
    api_group = "rbac.authorization.k8s.io"
    kind      = "ClusterRole"
    name      = kubernetes_cluster_role.argocd_server[0].metadata[0].name
  }

  subject {
    kind      = "ServiceAccount"
    name      = "argocd-server"
    namespace = kubernetes_namespace.argocd[0].metadata[0].name
  }
}
