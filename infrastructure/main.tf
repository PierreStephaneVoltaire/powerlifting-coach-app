locals {
  cluster_name = "${var.project_name}-${var.environment}"
}

# Resource Group
resource "azurerm_resource_group" "this" {
  name     = "${local.cluster_name}-rg"
  location = var.region
}

# Storage Account for blob storage (replaces Digital Ocean Spaces)
resource "azurerm_storage_account" "videos" {
  name                     = replace("${var.project_name}${var.environment}videos", "-", "")
  resource_group_name      = azurerm_resource_group.this.name
  location                 = azurerm_resource_group.this.location
  account_tier             = "Standard"
  account_replication_type = "LRS"
  account_kind             = "StorageV2"

  blob_properties {
    cors_rule {
      allowed_headers    = ["*"]
      allowed_methods    = ["GET", "HEAD"]
      allowed_origins    = ["*"]
      exposed_headers    = ["*"]
      max_age_in_seconds = 3600
    }
  }

  tags = {
    environment = var.environment
    project     = var.project_name
  }
}

# Storage Container for videos
resource "azurerm_storage_container" "videos" {
  name                  = var.storage_container_name
  storage_account_name  = azurerm_storage_account.videos.name
  container_access_type = "blob"
}

# Lifecycle Management Policy for blob expiration
resource "azurerm_storage_management_policy" "videos" {
  storage_account_id = azurerm_storage_account.videos.id

  rule {
    name    = "expire-after-120-days"
    enabled = true
    filters {
      blob_types = ["blockBlob"]
    }
    actions {
      base_blob {
        delete_after_days_since_modification_greater_than = 120
      }
    }
  }
}

# AKS cluster with spot instances
resource "azurerm_kubernetes_cluster" "k8s" {
  name                = local.cluster_name
  location            = azurerm_resource_group.this.location
  resource_group_name = azurerm_resource_group.this.name
  dns_prefix          = local.cluster_name
  kubernetes_version  = var.kubernetes_version

  default_node_pool {
    name                = "default"
    vm_size             = var.node_size
    enable_auto_scaling = true
    min_count           = 0
    max_count           = 3
    os_disk_size_gb     = 30

    # Enable spot instances for cost savings
    priority        = "Spot"
    eviction_policy = "Delete"
    spot_max_price  = -1 # Use Azure's current spot price
  }

  identity {
    type = "SystemAssigned"
  }

  network_profile {
    network_plugin = "azure"
    network_policy = "azure"
  }

  tags = {
    environment = var.environment
    project     = var.project_name
  }
}
