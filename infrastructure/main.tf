locals {
  cluster_name = "${var.project_name}-${var.environment}"
}

resource "azurerm_resource_group" "this" {
  name     = "${local.cluster_name}-rg"
  location = var.region
}

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

resource "azurerm_storage_container" "videos" {
  name                  = var.storage_container_name
  storage_account_name  = azurerm_storage_account.videos.name
  container_access_type = "blob"
}

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

resource "azurerm_kubernetes_cluster" "k8s" {
  name                = local.cluster_name
  location            = azurerm_resource_group.this.location
  resource_group_name = azurerm_resource_group.this.name
  dns_prefix          = local.cluster_name
  kubernetes_version  = var.kubernetes_version

  automatic_upgrade_channel = "stable"

  maintenance_window_auto_upgrade {
    frequency   = "Weekly"
    interval    = 1
    duration    = 4
    day_of_week = "Sunday"
    start_time  = "02:00"
    utc_offset  = "-05:00"
  }

  maintenance_window_node_os {
    frequency   = "Weekly"
    interval    = 1
    duration    = 4
    day_of_week = "Sunday"
    start_time  = "06:00"
    utc_offset  = "-05:00"
  }

  default_node_pool {
    name                 = "default"
    vm_size              = "Standard_B1s"
    auto_scaling_enabled = true
    min_count            = 1
    max_count            = 1
    os_disk_size_gb      = 30
    temporary_name_for_rotation = "defaulttemp"
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

resource "azurerm_kubernetes_cluster_node_pool" "spot" {
  name                  = "spot"
  kubernetes_cluster_id = azurerm_kubernetes_cluster.k8s.id
  vm_size               = var.spot_node_size
  auto_scaling_enabled = true
  min_count             = var.stopped ? 0 : var.spot_node_min_count
  max_count             = var.spot_node_max_count
  os_disk_size_gb       = 30
  priority              = "Spot"
  eviction_policy       = "Delete"
  spot_max_price        = -1


  node_labels = {
    "kubernetes.azure.com/scalesetpriority" = "spot"
    "workload-type"                         = "spot"
  }

  node_taints = [
    "kubernetes.azure.com/scalesetpriority=spot:NoSchedule"
  ]

  tags = {
    environment = var.environment
    project     = var.project_name
    node_type   = "spot"
  }
}
