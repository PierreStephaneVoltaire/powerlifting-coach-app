# Azure Communication Services for Email

locals {
  azure_email_smtp_host = "smtp.azurecomm.net"
}

resource "azurerm_email_communication_service" "this" {
  name                = "${var.project_name}-${var.environment}-email"
  resource_group_name = azurerm_resource_group.this.name
  data_location       = "United States"

  tags = {
    environment = var.environment
    project     = var.project_name
  }
}

resource "azurerm_email_communication_service_domain" "this" {
  name                = var.domain_name
  email_service_id    = azurerm_email_communication_service.this.id
  domain_management   = "CustomerManaged"

  tags = {
    environment = var.environment
    project     = var.project_name
  }
}

resource "azurerm_communication_service" "this" {
  name                = "${var.project_name}-${var.environment}-comm"
  resource_group_name = azurerm_resource_group.this.name
  data_location       = "United States"

  tags = {
    environment = var.environment
    project     = var.project_name
  }
}

# Link the email domain to the communication service
# This requires the domain to be verified first, so it's controlled by a variable
# Steps:
# 1. Apply with email_domain_verified=false (creates domain and DNS records)
# 2. Wait for Azure to verify the domain (check in Azure Portal)
# 3. Apply with email_domain_verified=true (links the domain)
resource "azurerm_communication_service_email_domain_association" "this" {
  count                    = var.email_domain_verified ? 1 : 0
  communication_service_id = azurerm_communication_service.this.id
  email_service_domain_id  = azurerm_email_communication_service_domain.this.id
}

# Output the connection string for SMTP
output "azure_email_connection_string" {
  description = "Azure Communication Services connection string (use as azure_email_smtp_password)"
  value       = azurerm_communication_service.this.primary_connection_string
  sensitive   = true
}

output "azure_email_smtp_username" {
  description = "SMTP username (your verified sender email)"
  value       = "noreply@${var.domain_name}"
}
