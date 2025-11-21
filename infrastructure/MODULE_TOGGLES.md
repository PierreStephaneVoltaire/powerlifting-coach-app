# Terraform Module Toggles

## Dependency Chain
```
rancher_cluster
├── kubernetes_base
│   └── kubernetes_monitoring (needs app_namespace, grafana_admin_password)
│   └── argocd_apps (needs app_namespace)
├── kubernetes_networking
├── argocd
│   └── argocd_apps (needs argocd_namespace)
```

---

## 0. Rancher Server Only (no modules)
```bash
terraform apply -target="aws_instance.rancher_server" -target="aws_eip.rancher" -target="aws_eip_association.rancher_server" -target="aws_route53_record.rancher_server" -target="aws_security_group.rancher_server" -target="aws_vpc.main" -target="aws_subnet.public" -target="aws_internet_gateway.main" -target="aws_route_table.public" -target="aws_route_table_association.public" -target="aws_key_pair.rancher" -target="aws_iam_role.rancher_server" -target="aws_iam_instance_profile.rancher_server" -target="aws_route53_zone.main"
```

## 1. Rancher Cluster Only
```bash
terraform apply -target="module.rancher_cluster"
```

## 2. Rancher + Kubernetes Base
```bash
terraform apply -target="module.rancher_cluster" -target="module.kubernetes_base"
```

## 3. Rancher + Kubernetes Base + Networking
```bash
terraform apply -target="module.rancher_cluster" -target="module.kubernetes_base" -target="module.kubernetes_networking"
```

## 4. Rancher + Kubernetes Base + Monitoring
```bash
terraform apply -target="module.rancher_cluster" -target="module.kubernetes_base" -target="module.kubernetes_monitoring"
```

## 5. Rancher + ArgoCD (no apps)
```bash
terraform apply -target="module.rancher_cluster" -target="module.argocd"
```

## 6. Rancher + Kubernetes Base + ArgoCD + ArgoCD Apps
```bash
terraform apply -target="module.rancher_cluster" -target="module.kubernetes_base" -target="module.argocd" -target="module.argocd_apps"
```

## 7. Full Stack (All Modules)
```bash
terraform apply -target="module.rancher_cluster" -target="module.kubernetes_base" -target="module.kubernetes_networking" -target="module.kubernetes_monitoring" -target="module.argocd" -target="module.argocd_apps"
```

---

## Destroy Combos (reverse order)

### Destroy ArgoCD Apps Only
```bash
terraform destroy -target="module.argocd_apps"
```

### Destroy ArgoCD + Apps
```bash
terraform destroy -target="module.argocd_apps" -target="module.argocd"
```

### Destroy Monitoring Only
```bash
terraform destroy -target="module.kubernetes_monitoring"
```

### Destroy Everything Except Rancher
```bash
terraform destroy -target="module.argocd_apps" -target="module.argocd" -target="module.kubernetes_monitoring" -target="module.kubernetes_networking" -target="module.kubernetes_base"
```

### Full Destroy
```bash
terraform destroy -target="module.argocd_apps" -target="module.argocd" -target="module.kubernetes_monitoring" -target="module.kubernetes_networking" -target="module.kubernetes_base" -target="module.rancher_cluster"
```
