output "nginx_gateway_fabric_name" {
  description = "NGINX Gateway Fabric release name"
  value       = var.stopped ? null : helm_release.nginx_gateway_fabric[0].name
}

output "letsencrypt_issuer_name" {
  description = "Let's Encrypt cluster issuer name"
  value       = kubectl_manifest.letsencrypt_prod.name
}
