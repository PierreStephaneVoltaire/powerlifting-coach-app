resource "random_password" "valkey_password" {
  length  = 32
  special = true
}

resource "random_password" "postgres_password" {
  length  = 32
  special = false
}



resource "kubernetes_secret" "valkey" {
  metadata {
    name      = "valkey"
    namespace = kubernetes_namespace.data.metadata[0].name
  }
  data = {
    password = random_password.valkey_password.result
    host     = "valkey.${kubernetes_namespace.data.metadata[0].name}.svc.cluster.local"
    port     = "6379"
  }
  type = "Opaque"
}


resource "kubernetes_secret" "postgres" {
  metadata {
    name      = "postgres"
    namespace = kubernetes_namespace.data.metadata[0].name
  }
  data = {
    username = "postgres"
    password = random_password.postgres_password.result
    host     = "postgres.${kubernetes_namespace.data.metadata[0].name}.svc.cluster.local"
    port     = "5432"
    database = "app"
  }
  type = "Opaque"
}
