cd stacks/argocd-apps
terraform state rm kubernetes_cluster_role_binding.argocd_server
terraform state rm kubernetes_service_account.argocd_app
terraform state rm kubernetes_role_binding.argocd_app
terraform state rm kubernetes_manifest.app_frontend[0]
terraform state rm kubernetes_manifest.app_datalayer[0]
terraform state rm kubernetes_manifest.app_backend[0]
terraform state rm kubernetes_cluster_role.argocd_server
terraform state rm kubernetes_role.argocd_app
cd ../argocd

terraform state rm  data.kubernetes_secret.argocd_admin[0]
terraform state rm  data.terraform_remote_state.rancher_cluster
terraform state rm  aws_ssm_parameter.argocd_admin_password[0]
terraform state rm  helm_release.argocd[0]
terraform state rm  kubectl_manifest.argocd_httproute
terraform state rm  kubernetes_namespace.argocd
cd ../kubernetes-monitoring



terraform state rm  data.terraform_remote_state.kubernetes_base
terraform state rm  data.terraform_remote_state.rancher_cluster
terraform state rm  helm_release.kube_prometheus_stack[0]
terraform state rm  helm_release.loki[0]
terraform state rm  helm_release.promtail[0]
terraform state rm  kubectl_manifest.grafana_httproute
terraform state rm  kubectl_manifest.loki_httproute
terraform state rm  kubectl_manifest.prometheus_httproute
terraform state rm  kubectl_manifest.rabbitmq_httproute
cd ../kubernetes-networking


terraform state rm  data.aws_route53_zone.main
terraform state rm  data.http.gateway_api_crds
terraform state rm  data.kubectl_file_documents.gateway_api_crds
terraform state rm  data.kubernetes_service.nginx_gateway[0]
terraform state rm  data.terraform_remote_state.rancher_cluster
terraform state rm  aws_route53_record.cluster_wildcard[0]
terraform state rm  helm_release.cert_manager
terraform state rm  helm_release.external_dns
terraform state rm  helm_release.nginx_gateway_fabric[0]
terraform state rm  kubectl_manifest.gateway_api_crds[0]
terraform state rm  kubectl_manifest.gateway_api_crds[1]
terraform state rm  kubectl_manifest.gateway_api_crds[2]
terraform state rm  kubectl_manifest.gateway_api_crds[3]
terraform state rm  kubectl_manifest.gateway_api_crds[4]
terraform state rm  kubectl_manifest.gateway_api_crds[5]
terraform state rm  kubectl_manifest.letsencrypt_prod
terraform state rm  kubectl_manifest.nginx_gateway[0]
cd ../kubernetes-base


terraform state rm  data.terraform_remote_state.base
terraform state rm  data.terraform_remote_state.rancher_cluster
terraform state rm  kubernetes_namespace.app
terraform state rm  kubernetes_secret.app_secrets
terraform state rm  kubernetes_secret.google_oauth_secret
terraform state rm  kubernetes_secret.keycloak_secret
terraform state rm  kubernetes_secret.postgres_secret
terraform state rm  kubernetes_secret.rabbitmq_secret
terraform state rm  kubernetes_secret.ses_email_secret
terraform state rm  local_file.auth_service_patch[0]
terraform state rm  local_file.frontend_patch[0]
terraform state rm  local_file.kustomization_patches[0]
terraform state rm  local_file.media_processor_service_patch[0]
terraform state rm  local_file.video_service_patch[0]
terraform state rm  random_password.grafana_admin_password
terraform state rm  random_password.keycloak_admin_password
terraform state rm  random_password.keycloak_client_secret
terraform state rm  random_password.postgres_password
terraform state rm  random_password.postgres_password
terraform state rm  random_password.rabbitmq_password
cd ../rancher-cluster


terraform apply -auto-approve
cd ../kubernetes-base
terraform apply -auto-approve
cd ../kubernetes-networking
terraform apply -auto-approve
cd ../kubernetes-monitoring
terraform apply -auto-approve
cd ../argocd
terraform apply -auto-approve
cd ../argocd-apps
terraform apply -auto-approve



