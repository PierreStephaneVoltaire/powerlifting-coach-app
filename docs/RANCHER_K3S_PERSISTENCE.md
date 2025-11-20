# Rancher and K3s Data Persistence Strategy

## Overview

This document outlines the data persistence and backup strategies for the Rancher management server and K3s cluster infrastructure.

## Critical Data Components

### 1. Rancher Server Data
**Location**: `/var/lib/rancher` on Rancher server EC2 instance
**Contains**:
- Rancher configuration database (SQLite or embedded DB)
- User accounts and RBAC settings
- Cluster registrations and configurations
- API keys and tokens
- Rancher catalogs and app deployments

### 2. K3s etcd Data
**Location**: `/var/lib/rancher/k3s/server/db/` on K3s nodes
**Contains**:
- All Kubernetes resources (Deployments, Services, ConfigMaps, Secrets)
- Cluster state and configuration
- Application deployments and configurations
- Critical for cluster recovery

## Infrastructure Setup

### Rancher Server Persistence

**EC2 Instance**: `t4g.nano` with Amazon Linux 2
**Storage Configuration**:
```
Root Volume:    30 GB GP3 (ephemeral, for OS)
Data Volume:    20 GB GP3 (persistent, for Rancher data)
  Device:       /dev/sdf (appears as /dev/nvme1n1)
  Mount Point:  /var/lib/rancher
  Filesystem:   ext4
  Encrypted:    Yes
```

**Docker Volume Mount**:
```bash
docker run -d --restart=unless-stopped \
  -v /var/lib/rancher:/var/lib/rancher \  # ← Persistent EBS volume
  -v /opt/rancher/ssl:/etc/rancher/ssl:ro \
  rancher/rancher:latest
```

### K3s Cluster Nodes

**Instance Type**: `t3a.small` spot instances
**Node Roles**: Control plane + etcd + worker (all-in-one)
**Root Volume**: 30 GB (ephemeral)

## Backup Strategy

### Rancher Server Backups

#### 1. EBS Volume Snapshots

**Automatic snapshots** via AWS Data Lifecycle Manager:
- **Frequency**: Daily at 3 AM UTC
- **Retention**: 7 days
- **Target**: EBS volume tagged with `Backup=Daily`

**Manual snapshot for major changes**:
```bash
aws ec2 create-snapshot \
  --volume-id vol-xxxxx \
  --description "Pre-upgrade Rancher backup $(date +%Y%m%d)"
```

#### 2. Rancher Database Export

For additional safety, export Rancher's database:
```bash
# SSH into Rancher server
ssh -i rancher-key.pem ec2-user@rancher.yourdomain.com

# Stop Rancher container
docker stop $(docker ps -q --filter ancestor=rancher/rancher:latest)

# Create backup
sudo tar czf rancher-backup-$(date +%Y%m%d).tar.gz /var/lib/rancher

# Upload to S3
aws s3 cp rancher-backup-*.tar.gz s3://your-backup-bucket/rancher/

# Restart Rancher
docker start $(docker ps -aq --filter ancestor=rancher/rancher:latest)
```

### K3s etcd Backups

#### Automatic Snapshots to S3

**Configuration** (infrastructure/rancher-cluster.tf:224-233):
```hcl
etcd {
  snapshot_schedule_cron = "0 */6 * * *"  # Every 6 hours
  snapshot_retention     = 10              # Keep 10 snapshots
  s3_config {
    bucket    = "powerlifting-coach-etcd-backups"
    region    = "us-east-1"
    folder    = "etcd-snapshots"
  }
}
```

**Backup Schedule**:
- Every 6 hours (00:00, 06:00, 12:00, 18:00 UTC)
- Keeps last 10 snapshots (~2.5 days)
- Automatically uploaded to S3
- S3 lifecycle policy deletes after 7 days

**S3 Bucket Features**:
- ✅ Versioning enabled
- ✅ Encryption at rest (AES256)
- ✅ Public access blocked
- ✅ Lifecycle policy for cleanup

## Data Loss Scenarios

### Scenario 1: Rancher Container Restart
**Risk**: LOW
**Protection**: EBS volume persists, data intact
**Recovery Time**: 2-3 minutes (container restart)
**Data Loss**: None

### Scenario 2: Rancher EC2 Instance Termination
**Risk**: MEDIUM
**Protection**: EBS volume survives (not deleted with instance)
**Recovery Steps**:
1. Launch new EC2 instance in same AZ
2. Stop instance, detach root volume
3. Attach existing Rancher data volume
4. Modify /etc/fstab to mount volume
5. Start Rancher container
**Recovery Time**: 15-30 minutes
**Data Loss**: None if volume is healthy

### Scenario 3: Single K3s Node Failure (Spot Termination)
**Risk**: HIGH (spot instances)
**Protection**: etcd snapshots in S3
**Impact**: If 1 node fails and you have 3+ nodes, cluster continues
**Data Loss**: None (etcd is replicated)

### Scenario 4: All K3s Nodes Lost Simultaneously
**Risk**: VERY HIGH (all spot instances terminated)
**Protection**: S3 etcd snapshots
**Recovery Steps**:
1. Create new cluster via Rancher UI
2. Download latest etcd snapshot from S3
3. Restore etcd data to new cluster
4. Verify all resources are present
**Recovery Time**: 1-2 hours
**Data Loss**: Up to 6 hours (last snapshot interval)

### Scenario 5: Rancher EBS Volume Corruption
**Risk**: LOW but CATASTROPHIC
**Protection**: Daily EBS snapshots
**Recovery Steps**:
1. Create new EBS volume from latest snapshot
2. Attach to Rancher instance
3. Restart Rancher container
**Recovery Time**: 30 minutes
**Data Loss**: Up to 24 hours (last snapshot)

## Disaster Recovery Procedures

### Restoring Rancher Server

#### From EBS Snapshot
```bash
# 1. Create volume from snapshot
aws ec2 create-volume \
  --snapshot-id snap-xxxxx \
  --availability-zone us-east-1a \
  --volume-type gp3

# 2. Attach to instance
aws ec2 attach-volume \
  --volume-id vol-yyyyy \
  --instance-id i-zzzzz \
  --device /dev/sdf

# 3. SSH and mount
ssh -i rancher-key.pem ec2-user@rancher-ip
sudo mount /dev/nvme1n1 /var/lib/rancher

# 4. Restart Rancher container
docker restart $(docker ps -q --filter ancestor=rancher/rancher:latest)
```

#### From Manual Backup
```bash
# 1. Download backup from S3
aws s3 cp s3://backup-bucket/rancher/rancher-backup-20250315.tar.gz .

# 2. Extract
sudo tar xzf rancher-backup-20250315.tar.gz -C /

# 3. Set permissions
sudo chown -R 1000:1000 /var/lib/rancher

# 4. Restart Rancher
docker restart <container-id>
```

### Restoring K3s etcd

#### From Automatic S3 Snapshot
```bash
# 1. List available snapshots
aws s3 ls s3://powerlifting-coach-etcd-backups/etcd-snapshots/

# 2. Through Rancher UI:
#    - Go to Cluster Management
#    - Select cluster
#    - Click "Restore Snapshot"
#    - Choose snapshot from S3
#    - Confirm restoration

# Or via CLI (on K3s node):
# 3. Download snapshot
aws s3 cp s3://bucket/etcd-snapshots/etcd-snapshot-xxx .

# 4. Stop K3s
sudo systemctl stop k3s

# 5. Restore
sudo k3s server \
  --cluster-reset \
  --cluster-reset-restore-path=./etcd-snapshot-xxx

# 6. Start K3s
sudo systemctl start k3s
```

## Prevention and Best Practices

### Before Major Changes

**Always take manual backups before**:
- Upgrading Rancher
- Upgrading Kubernetes version
- Making RBAC changes
- Adding/removing clusters

```bash
# Rancher backup
docker exec $(docker ps -q --filter ancestor=rancher/rancher) \
  sqlite3 /var/lib/rancher/management-state/management-state.db \
  ".backup '/tmp/rancher-pre-upgrade.db'"

# etcd snapshot (from K3s node)
kubectl --kubeconfig=/etc/rancher/k3s/k3s.yaml \
  exec -n kube-system etcd-xxx -- \
  etcdctl snapshot save /tmp/pre-upgrade-snapshot.db
```

### Regular Testing

**Monthly DR Drill**:
1. Document current cluster state
2. Restore from latest backup to test cluster
3. Verify all resources present
4. Document recovery time
5. Update this document with findings

### Monitoring

**Set up alerts for**:
- Etcd snapshot failures
- S3 upload failures
- EBS snapshot failures
- Rancher container health
- K3s node availability

**CloudWatch Alarms**:
```bash
# Etcd backup failure alarm
aws cloudwatch put-metric-alarm \
  --alarm-name etcd-backup-failure \
  --metric-name SnapshotFailures \
  --threshold 1 \
  --comparison-operator GreaterThanThreshold
```

## Spot Instance Considerations

### Risk Mitigation

Since K3s nodes use **spot instances**, they can be terminated with 2-minute notice:

1. **Mixed instance types**: Consider adding on-demand instances for control plane
2. **Multiple AZs**: Spread nodes across availability zones
3. **Frequent backups**: 6-hour interval balances cost vs RPO
4. **Spot instance interruption handler**: Consider AWS Node Termination Handler

### Recommended Architecture Change

For production, consider:
```hcl
# Control plane nodes (on-demand)
machine_pools {
  name               = "control-plane"
  control_plane_role = true
  etcd_role         = true
  worker_role       = false
  quantity          = 3
  # instance_type = "t3a.small" (on-demand)
}

# Worker nodes (spot)
machine_pools {
  name               = "workers"
  control_plane_role = false
  etcd_role         = false
  worker_role       = true
  quantity          = 2
  # spot instances OK for workers
}
```

## Cost Analysis

**Current Setup**:
- Rancher EBS data volume: $2/month (20 GB GP3)
- EBS snapshots: ~$0.50/month (7 daily snapshots)
- etcd S3 storage: ~$0.10/month (small files)
- S3 PUT/GET requests: ~$0.05/month

**Total**: ~$2.65/month for complete backup coverage

**Optional Improvements**:
- On-demand control plane nodes: +$15-20/month
- Multi-AZ deployment: +$5/month (data transfer)
- More frequent snapshots: +$0.02/month per snapshot

## Migration to High Availability

### Recommended HA Setup

**Rancher Server**:
1. Deploy Rancher on K3s cluster (instead of Docker)
2. Use RDS PostgreSQL for Rancher database
3. Run 3 replicas behind load balancer
4. Store SSL certs in AWS Secrets Manager

**K3s Cluster**:
1. 3 control plane nodes (on-demand) in different AZs
2. 2+ worker nodes (spot is OK)
3. External etcd cluster OR embedded etcd with 3+ nodes
4. AWS ELB for API server
5. EBS CSI driver for persistent workloads

**Cost**: ~$60-80/month (much more reliable)

## Backup Verification

### Weekly Verification Script

```bash
#!/bin/bash
# verify-backups.sh

# Check Rancher EBS snapshots
LATEST_SNAP=$(aws ec2 describe-snapshots \
  --owner-ids self \
  --filters "Name=tag:Name,Values=*rancher-data*" \
  --query 'Snapshots | sort_by(@, &StartTime) | [-1].SnapshotId' \
  --output text)

SNAP_AGE=$(aws ec2 describe-snapshots \
  --snapshot-ids $LATEST_SNAP \
  --query 'Snapshots[0].StartTime' \
  --output text)

echo "Latest Rancher snapshot: $LATEST_SNAP from $SNAP_AGE"

# Check etcd backups in S3
LATEST_ETCD=$(aws s3 ls s3://powerlifting-coach-etcd-backups/etcd-snapshots/ \
  | sort | tail -n 1 | awk '{print $4}')

echo "Latest etcd snapshot: $LATEST_ETCD"

# Alert if backups are too old
# Add logic to send SNS notification if backups are > 24h old
```

## Recovery Time Objectives (RTO) and Recovery Point Objectives (RPO)

| Scenario | RTO | RPO | Priority |
|----------|-----|-----|----------|
| Rancher container restart | 5 min | 0 | High |
| Rancher instance failure | 30 min | 0 | High |
| Single K3s node failure | 10 min | 0 | Medium |
| Multiple K3s nodes lost | 2 hours | 6 hours | High |
| Complete cluster loss | 4 hours | 6 hours | Critical |
| Rancher data corruption | 1 hour | 24 hours | Critical |

## Automation Recommendations

1. **Terraform state backup**: Store in S3 with versioning
2. **Automated DR testing**: Monthly Lambda function to test restore
3. **Backup monitoring**: CloudWatch + SNS notifications
4. **Documentation in code**: Keep this doc updated with Terraform changes

## Emergency Contacts

- AWS Support: [Your support plan]
- On-call Engineer: [Your contact]
- Backup Verification Dashboard: [Link to monitoring]

## Changelog

- 2025-03-15: Initial implementation of EBS persistence for Rancher
- 2025-03-15: Added etcd S3 snapshots every 6 hours
- 2025-03-15: Configured S3 lifecycle policies for backup retention
