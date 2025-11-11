resource "azurerm_dns_zone" "main" {
  count               = var.domain_name != "localhost" ? 1 : 0
  name                = var.domain_name
  resource_group_name = azurerm_resource_group.this.name

  tags = {
    environment = var.environment
    project     = var.project_name
  }
}

locals {
  dns_lb_ip = var.kubernetes_resources_enabled && var.domain_name != "localhost" ? (
    length(data.kubernetes_service.nginx_ingress) > 0 ?
    data.kubernetes_service.nginx_ingress[0].status[0].load_balancer[0].ingress[0].ip : null
  ) : null
}

resource "azurerm_dns_a_record" "app" {
  count               = var.domain_name != "localhost" && var.kubernetes_resources_enabled ? 1 : 0
  name                = "app"
  zone_name           = azurerm_dns_zone.main[0].name
  resource_group_name = azurerm_resource_group.this.name
  ttl                 = 300
  records             = [local.dns_lb_ip]

  tags = {
    environment = var.environment
    project     = var.project_name
  }
}

resource "azurerm_dns_a_record" "api" {
  count               = var.domain_name != "localhost" && var.kubernetes_resources_enabled ? 1 : 0
  name                = "api"
  zone_name           = azurerm_dns_zone.main[0].name
  resource_group_name = azurerm_resource_group.this.name
  ttl                 = 300
  records             = [local.dns_lb_ip]

  tags = {
    environment = var.environment
    project     = var.project_name
  }
}

resource "azurerm_dns_a_record" "auth" {
  count               = var.domain_name != "localhost" && var.kubernetes_resources_enabled ? 1 : 0
  name                = "auth"
  zone_name           = azurerm_dns_zone.main[0].name
  resource_group_name = azurerm_resource_group.this.name
  ttl                 = 300
  records             = [local.dns_lb_ip]

  tags = {
    environment = var.environment
    project     = var.project_name
  }
}

resource "azurerm_dns_txt_record" "email_verification" {
  count               = var.domain_name != "localhost" && var.azure_email_domain_verification_code != "" ? 1 : 0
  name                = "@"
  zone_name           = azurerm_dns_zone.main[0].name
  resource_group_name = azurerm_resource_group.this.name
  ttl                 = 300

  record {
    value = var.azure_email_domain_verification_code
  }

  tags = {
    environment = var.environment
    project     = var.project_name
  }
}

resource "azurerm_dns_txt_record" "spf" {
  count               = var.domain_name != "localhost" && var.kubernetes_resources_enabled ? 1 : 0
  name                = "@"
  zone_name           = azurerm_dns_zone.main[0].name
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

resource "azurerm_dns_mx_record" "email" {
  count               = var.domain_name != "localhost" && var.enable_mx_records ? 1 : 0
  name                = "@"
  zone_name           = azurerm_dns_zone.main[0].name
  resource_group_name = azurerm_resource_group.this.name
  ttl                 = 300

  record {
    preference = 10
    exchange   = var.azure_email_mx_endpoint
  }

  tags = {
    environment = var.environment
    project     = var.project_name
  }
}

resource "azurerm_dns_txt_record" "dkim1" {
  count               = var.domain_name != "localhost" && var.azure_email_dkim_selector1 != "" ? 1 : 0
  name                = var.azure_email_dkim_selector1
  zone_name           = azurerm_dns_zone.main[0].name
  resource_group_name = azurerm_resource_group.this.name
  ttl                 = 300

  record {
    value = var.azure_email_dkim_value1
  }

  tags = {
    environment = var.environment
    project     = var.project_name
  }
}

resource "azurerm_dns_txt_record" "dkim2" {
  count               = var.domain_name != "localhost" && var.azure_email_dkim_selector2 != "" ? 1 : 0
  name                = var.azure_email_dkim_selector2
  zone_name           = azurerm_dns_zone.main[0].name
  resource_group_name = azurerm_resource_group.this.name
  ttl                 = 300

  record {
    value = var.azure_email_dkim_value2
  }

  tags = {
    environment = var.environment
    project     = var.project_name
  }
}


resource "azurerm_dns_txt_record" "dmarc" {
  count               = var.domain_name != "localhost" && var.kubernetes_resources_enabled ? 1 : 0
  name                = "_dmarc"
  zone_name           = azurerm_dns_zone.main[0].name
  resource_group_name = azurerm_resource_group.this.name
  ttl                 = 300

  record {
    value = "v=DMARC1; p=quarantine; rua=mailto:${var.azure_email_from_email}"
  }

  tags = {
    environment = var.environment
    project     = var.project_name
  }
}
