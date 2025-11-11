# Azure Communication Services for Email

locals {
  azure_email_smtp_host = "smtp.azurecomm.net"
}

resource "azurerm_email_communication_service" "this" {
  count               = var.domain_name != "localhost" ? 1 : 0
  name                = "${var.project_name}-${var.environment}-email"
  resource_group_name = azurerm_resource_group.this.name
  data_location       = "United States"

  tags = {
    environment = var.environment
    project     = var.project_name
  }
}

resource "azurerm_email_communication_service_domain" "this" {
  count               = var.domain_name != "localhost" ? 1 : 0
  name                = var.domain_name
  email_service_id    = azurerm_email_communication_service.this[0].id
  domain_management   = "CustomerManaged"

  tags = {
    environment = var.environment
    project     = var.project_name
  }
}

resource "azurerm_communication_service" "this" {
  count               = var.domain_name != "localhost" ? 1 : 0
  name                = "${var.project_name}-${var.environment}-comm"
  resource_group_name = azurerm_resource_group.this.name
  data_location       = "United States"

  tags = {
    environment = var.environment
    project     = var.project_name
  }
}

# Link the email domain to the communication service
resource "azurerm_communication_service_email_domain_association" "this" {
  count                  = var.domain_name != "localhost" ? 1 : 0
  communication_service_id = azurerm_communication_service.this[0].id
  email_service_domain_id  = azurerm_email_communication_service_domain.this[0].id
}

# Output the connection string for SMTP
output "azure_email_connection_string" {
  description = "Azure Communication Services connection string (use as azure_email_smtp_password)"
  value       = var.domain_name != "localhost" ? azurerm_communication_service.this[0].primary_connection_string : "not-configured"
  sensitive   = true
}

output "azure_email_smtp_username" {
  description = "SMTP username (your verified sender email)"
  value       = var.domain_name != "localhost" ? "noreply@${var.domain_name}" : "not-configured"
}
