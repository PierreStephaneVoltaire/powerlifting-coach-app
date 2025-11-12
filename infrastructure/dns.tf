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

# Combined TXT record for @ - includes both SPF and email verification
resource "azurerm_dns_txt_record" "root" {
  name                = "@"
  zone_name           = azurerm_dns_zone.main.name
  resource_group_name = azurerm_resource_group.this.name
  ttl                 = 300

  # SPF record
  record {
    value = "v=spf1 include:azurecomm.net ~all"
  }

  # Email domain verification
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

# DKIM CNAME records for email authentication
resource "azurerm_dns_cname_record" "dkim1" {
  name                = "selector1-azurecomm-prod-net._domainkey"
  zone_name           = azurerm_dns_zone.main.name
  resource_group_name = azurerm_resource_group.this.name
  ttl                 = 3600
  record              = "selector1-azurecomm-prod-net._domainkey.azurecomm.net"

  tags = {
    environment = var.environment
    project     = var.project_name
  }

  depends_on = [
    azurerm_email_communication_service_domain.this
  ]
}

resource "azurerm_dns_cname_record" "dkim2" {
  name                = "selector2-azurecomm-prod-net._domainkey"
  zone_name           = azurerm_dns_zone.main.name
  resource_group_name = azurerm_resource_group.this.name
  ttl                 = 3600
  record              = "selector2-azurecomm-prod-net._domainkey.azurecomm.net"

  tags = {
    environment = var.environment
    project     = var.project_name
  }

  depends_on = [
    azurerm_email_communication_service_domain.this
  ]
}

resource "azurerm_dns_txt_record" "dmarc" {
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
