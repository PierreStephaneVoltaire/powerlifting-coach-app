resource "azurerm_dns_zone" "main" {
  name                = var.domain_name
  resource_group_name = azurerm_resource_group.this.name

  tags = {
    environment = var.environment
    project     = var.project_name
  }
}

locals {
  dns_lb_ip = var.kubernetes_resources_enabled && !var.stopped ? (
    length(data.kubernetes_service.nginx_ingress) > 0 ?
    data.kubernetes_service.nginx_ingress[0].status[0].load_balancer[0].ingress[0].ip : null
  ) : null
}

locals {
  subdomains = toset([
    "app",
    "api",
    "auth",
    "argocd",
    "grafana",
    "prometheus",
    "loki",
    "rabbitmq",
    "openwebui"
  ])
}

resource "azurerm_dns_a_record" "subdomains" {
  for_each            = var.kubernetes_resources_enabled && !var.stopped ? local.subdomains : []
  name                = each.key
  zone_name           = azurerm_dns_zone.main.name
  resource_group_name = azurerm_resource_group.this.name
  ttl                 = 300
  records             = [local.dns_lb_ip]

  tags = {
    environment = var.environment
    project     = var.project_name
    subdomain   = each.key
  }
}

resource "azurerm_dns_txt_record" "email_verification" {
  name                = "@"
  zone_name           = azurerm_dns_zone.main.name
  resource_group_name = azurerm_resource_group.this.name
  ttl                 = 300

  record {
    value = azurerm_email_communication_service_domain.this.verification_records[0].domain[0].value
  }

  tags = {
    environment = var.environment
    project     = var.project_name
  }

  depends_on = [
    azurerm_email_communication_service_domain.this
  ]
}

resource "azurerm_dns_txt_record" "spf" {
  count               = var.kubernetes_resources_enabled ? 1 : 0
  name                = "@"
  zone_name           = azurerm_dns_zone.main.name
  resource_group_name = azurerm_resource_group.this.name
  ttl                 = 300

  record {
    value = "v=spf1 include:azurecomm.net ~all"
  }

  tags = {
    environment = var.environment
    project     = var.project_name
  }
}

resource "azurerm_dns_txt_record" "dmarc" {
  count               = var.kubernetes_resources_enabled ? 1 : 0
  name                = "_dmarc"
  zone_name           = azurerm_dns_zone.main.name
  resource_group_name = azurerm_resource_group.this.name
  ttl                 = 300

  record {
    value = "v=DMARC1; p=quarantine; rua=mailto:noreply@${var.domain_name}"
  }

  tags = {
    environment = var.environment
    project     = var.project_name
  }
}
