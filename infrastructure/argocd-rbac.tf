resource "kubernetes_service_account" "argocd_app" {
  metadata {
    name      = "argocd-application-controller"
    namespace = kubernetes_namespace.app.metadata[0].name
  }
}

resource "kubernetes_role" "argocd_app" {
  metadata {
    name      = "argocd-application-controller"
    namespace = kubernetes_namespace.app.metadata[0].name
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
    namespace = kubernetes_namespace.app.metadata[0].name
  }

  role_ref {
    api_group = "rbac.authorization.k8s.io"
    kind      = "Role"
    name      = kubernetes_role.argocd_app.metadata[0].name
  }

  subject {
    kind      = "ServiceAccount"
    name      = "argocd-application-controller"
    namespace = kubernetes_namespace.argocd.metadata[0].name
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
    namespace = kubernetes_namespace.argocd.metadata[0].name
  }
}
