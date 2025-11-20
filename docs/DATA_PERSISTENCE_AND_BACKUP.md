# Data Persistence and Backup Strategy

## Overview

This document outlines the data persistence architecture and backup strategies for the Powerlifting Coach application.

## Current Data Persistence Setup

### Persistent Volume Claims (PVCs)

All stateful services now use AWS EBS GP3 volumes for persistent storage:

1. **PostgreSQL** - 10Gi EBS volume
   - Stores: Application data, user data, and Keycloak authentication data
   - Storage Class: `ebs-gp3`
   - Reclaim Policy: Retain (volumes are NOT deleted when PVC is deleted)

2. **RabbitMQ** - 5Gi EBS volume
   - Stores: Message queues and durable messages
   - Storage Class: `ebs-gp3`
   - Reclaim Policy: Retain

3. **Valkey (Redis)** - 2Gi EBS volume
   - Stores: Cache and session data
   - Storage Class: `ebs-gp3`
   - Reclaim Policy: Retain
   - AOF (Append-Only File) enabled for persistence

### Storage Class Configuration

The `ebs-gp3` StorageClass is configured with:
- **Type**: GP3 (General Purpose SSD)
- **Encryption**: Enabled
- **IOPS**: 3000 (baseline)
- **Throughput**: 125 MiB/s
- **Volume Binding**: WaitForFirstConsumer (volume created in same AZ as pod)
- **Expansion**: Allowed (can resize volumes without downtime)
- **Reclaim Policy**: Retain (prevents accidental data loss)

## Why EBS Volumes?

Previously, the system used `local-path` storage, which has critical limitations:

❌ **Problems with local-path**:
- Data stored on EC2 instance's local disk
- If node/instance fails → data is lost
- If pod reschedules to different node → can't access data
- **CRITICAL**: Using spot instances means AWS can terminate at any time

✅ **Benefits of EBS volumes**:
- Persist independently of EC2 instances
- Automatically replicated within availability zone
- Can be attached to different nodes if pods reschedule
- Survive node termination/replacement
- Support snapshots for backups
- Can be expanded without downtime

## Data Loss Scenarios and Protections

### Scenario 1: Pod Restart
- **Risk**: LOW
- **Protection**: EBS volume persists, pod remounts same volume
- **Data Loss**: None

### Scenario 2: Node Failure
- **Risk**: LOW
- **Protection**: Kubernetes reschedules pod to healthy node, attaches existing EBS volume
- **Data Loss**: None (brief downtime during rescheduling)

### Scenario 3: Spot Instance Termination
- **Risk**: MEDIUM (frequent with spot instances)
- **Protection**: EBS volume persists, pod reschedules with same volume
- **Data Loss**: None (5-10 minute downtime)

### Scenario 4: Accidental PVC Deletion
- **Risk**: HIGH (operator error)
- **Protection**: `Retain` reclaim policy - EBS volume is NOT deleted
- **Data Loss**: None if volume manually reattached
- **Recovery**: Requires manual intervention to create new PVC bound to existing volume

### Scenario 5: EBS Volume Corruption
- **Risk**: LOW but CATASTROPHIC
- **Protection**: Requires regular backups
- **Data Loss**: All data since last backup

### Scenario 6: Availability Zone Failure
- **Risk**: VERY LOW but CATASTROPHIC
- **Protection**: Requires cross-region backups
- **Data Loss**: All data since last backup

## Backup Strategy

### Recommended Approach

#### 1. PostgreSQL Backups (CRITICAL)

**Option A: pg_dump via CronJob (Recommended for now)**
```bash
# Daily backups to S3
kubectl create cronjob postgres-backup \
  --image=postgres:15-alpine \
  --schedule="0 2 * * *" \
  --restart=Never \
  -- /bin/sh -c "pg_dump -U app_user -h postgres powerlifting_app | gzip | aws s3 cp - s3://your-backup-bucket/postgres/backup-\$(date +%Y%m%d-%H%M%S).sql.gz"
```

**Option B: AWS Backup (Enterprise)**
- Create backup plan in AWS Backup service
- Tag EBS volumes for automatic backup
- Set retention policy (7 daily, 4 weekly, 12 monthly)

**Option C: Migrate to RDS (Strongly Recommended)**
- Move to Amazon RDS PostgreSQL
- Automatic daily backups with 7-35 day retention
- Point-in-time recovery (PITR)
- Multi-AZ deployment for high availability
- Automated patching and maintenance

#### 2. RabbitMQ Backups (MEDIUM Priority)

Message queues should be designed to be transient. However:
- Export definitions periodically
- Store in version control
- Test recovery procedures

#### 3. Valkey Backups (LOW Priority)

Cache data can be regenerated:
- If using for sessions: Consider JWT tokens instead
- If using for cache: Data can be recomputed
- AOF persistence already enabled for critical data

#### 4. EBS Snapshots

Create daily snapshots of all EBS volumes:
```bash
# Tag volumes for automatic snapshots
aws ec2 create-tags \
  --resources <volume-id> \
  --tags Key=Backup,Value=Daily

# Create snapshot lifecycle policy
aws dlm create-lifecycle-policy \
  --execution-role-arn <role-arn> \
  --description "Daily EBS snapshots" \
  --state ENABLED \
  --policy-details file://snapshot-policy.json
```

### Backup Schedule

| Data Type | Frequency | Retention | Method |
|-----------|-----------|-----------|--------|
| PostgreSQL | Daily (2 AM) | 30 days | pg_dump to S3 |
| EBS Snapshots | Daily (3 AM) | 7 days | AWS DLM |
| Application Code | On commit | Indefinite | Git |
| K8s Manifests | On commit | Indefinite | Git |

### Disaster Recovery Testing

**Monthly**: Test backup restoration:
1. Create test namespace
2. Restore latest PostgreSQL backup
3. Verify data integrity
4. Document recovery time (RTO/RPO)

**Quarterly**: Full DR drill:
1. Simulate complete cluster failure
2. Restore from backups to new cluster
3. Verify application functionality
4. Update DR documentation

## Migration Path: Moving to Managed Services

### PostgreSQL → RDS

**Benefits**:
- No manual backup management
- Automatic failover (Multi-AZ)
- Better performance and scalability
- Reduced operational burden

**Steps**:
1. Create RDS instance in same VPC
2. Configure security groups
3. Dump data from K8s PostgreSQL
4. Import to RDS
5. Update service connection strings
6. Test thoroughly
7. Decommission K8s PostgreSQL

**Cost Consideration**: ~$30-50/month for db.t3.small Multi-AZ

### RabbitMQ → Amazon MQ

**Benefits**:
- Managed service
- Automatic failover
- No maintenance burden

**Alternative**: Consider replacing with SNS/SQS for simpler architecture

## Monitoring and Alerts

### Critical Metrics to Monitor

1. **EBS Volume Usage**:
   - Alert at 80% capacity
   - Auto-expand if possible

2. **Backup Job Success**:
   - Alert on failed backups
   - Alert on missing backups

3. **PostgreSQL Health**:
   - Replication lag (if using replicas)
   - Connection pool exhaustion
   - Slow queries

4. **RabbitMQ**:
   - Queue depth
   - Message acknowledgment rate
   - Memory usage

## Cost Optimization

Current storage costs (approximate):
- PostgreSQL EBS (10 GB): ~$1/month
- RabbitMQ EBS (5 GB): ~$0.50/month
- Valkey EBS (2 GB): ~$0.20/month
- Snapshots (17 GB × 7 days): ~$0.85/month

**Total**: ~$2.55/month for persistent storage

## Emergency Procedures

### If Data Loss Occurs

1. **Stop all write operations immediately**
2. Identify the scope of data loss
3. Check if EBS volume still exists (reclaim policy = Retain)
4. If volume exists:
   - Create new PVC with existing volume ID
   - Verify data integrity
5. If volume lost:
   - Restore from latest backup
   - Calculate data loss window
   - Notify stakeholders

### Contact Information

- AWS Support: [Include your support plan details]
- On-call Engineer: [Include contact]
- Backup Verification Dashboard: [Include link]

## Next Steps

1. ✅ Implement EBS-backed PVCs (COMPLETED)
2. ⬜ Set up PostgreSQL backup CronJob
3. ⬜ Configure EBS snapshot lifecycle policy
4. ⬜ Test backup restoration procedure
5. ⬜ Evaluate RDS migration
6. ⬜ Set up monitoring and alerting
7. ⬜ Document and practice DR procedures

## References

- [AWS EBS Documentation](https://docs.aws.amazon.com/ebs/)
- [Kubernetes Persistent Volumes](https://kubernetes.io/docs/concepts/storage/persistent-volumes/)
- [PostgreSQL Backup Best Practices](https://www.postgresql.org/docs/current/backup.html)
