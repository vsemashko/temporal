# Disaster Recovery Runbook

**Version:** 1.0
**Last Updated:** 2025-11-22
**Owner:** SRE Team
**Status:** Ready for Testing

---

## Executive Summary

This runbook provides step-by-step procedures for responding to disaster scenarios affecting Temporal Server deployments. It covers detection, response, recovery, and verification procedures for various failure scenarios.

**Recovery Objectives:**
- **RTO (Recovery Time Objective):** <5 minutes
- **RPO (Recovery Point Objective):** <1 minute
- **Data Consistency:** 100% (no data loss)
- **Workflow Continuity:** 100% (all workflows resume)

**Emergency Contacts:**
- On-Call SRE: [pagerduty/opsgenie rotation]
- Security Team: security@example.com
- Platform Lead: platform-lead@example.com
- Executive Escalation: exec-oncall@example.com

---

## Quick Links (Emergency)

**STOP!** If you're responding to an active incident, jump to:

- [🔥 Region Failure](#scenario-1-single-region-failure) - Entire region down
- [💾 Database Failure](#scenario-2-database-failure) - Database unavailable
- [🔐 Security Incident](#scenario-3-security-breach) - Potential compromise
- [⚡ Cascading Failure](#scenario-4-cascading-failure) - Multiple components failing
- [🌐 Network Partition](#scenario-5-network-partition) - Split-brain scenario

**Pre-Incident Preparation:**
- Verify access to this runbook (bookmark it!)
- Ensure you have credentials for all systems
- Test failover procedures quarterly
- Review incident command structure

---

## Incident Command Structure

### Roles & Responsibilities

**Incident Commander (IC)**
- Owns the incident response
- Makes final decisions on mitigation strategies
- Coordinates communication
- Declares incident resolved

**Technical Lead (TL)**
- Executes technical recovery procedures
- Reports status to IC
- Requests additional resources

**Communications Lead (CL)**
- Updates stakeholders
- Maintains incident log
- Coordinates with customer support

**Scribe**
- Documents all actions taken
- Timestamps decision points
- Creates incident timeline

---

## Scenario 1: Single Region Failure

### Description
Primary region becomes unavailable due to datacenter failure, network outage, or zone-wide incident.

### Detection

**Automated Alerts:**
```promql
# Primary alert
up{job="temporal", region="primary"} == 0 for 1 minute

# Confirm with:
temporal_cluster_health{region="primary"} < 1
```

**Manual Verification:**
```bash
# Check service health
curl -f https://temporal-primary.example.com:7243/health || echo "PRIMARY DOWN"

# Check from secondary region
curl -f https://temporal-secondary.example.com:7243/health || echo "SECONDARY DOWN"

# Verify database connectivity
psql -h db-primary.example.com -U temporal -c "SELECT 1"
```

**Symptoms:**
- All Temporal services in primary region unreachable
- Workflows not progressing
- Workers unable to poll for tasks
- Client SDK connection failures

### Response Procedure

#### Phase 1: Verify Failure (Target: 1 minute)

```bash
# 1. Verify region is actually down (not just monitoring issue)
./scripts/verify_region_health.sh primary

# 2. Check blast radius
./scripts/check_affected_namespaces.sh primary

# 3. Verify secondary region healthy
./scripts/verify_region_health.sh secondary

# 4. Check replication lag
./scripts/check_replication_lag.sh
```

**Decision Point:** If replication lag >1 minute, PAUSE and escalate to Platform Lead.

#### Phase 2: Initiate Failover (Target: 2 minutes)

```bash
# 1. Declare incident
./scripts/incident_declare.sh --type=region-failure --region=primary --severity=critical

# 2. Put primary region in maintenance mode (if accessible)
temporal operator cluster mark-degraded --cluster=primary --reason="Region failure" || true

# 3. Promote secondary to primary
temporal operator cluster failover \
  --source-cluster=primary \
  --target-cluster=secondary \
  --force

# 4. Update DNS/Load Balancer
./scripts/update_dns.sh --primary=secondary --ttl=60

# 5. Verify DNS propagation
for i in {1..30}; do
  dig temporal.example.com +short
  sleep 2
done
```

#### Phase 3: Verify Recovery (Target: 1 minute)

```bash
# 1. Verify workflows resuming
temporal workflow list --namespace production --address temporal.example.com:7233

# 2. Check workflow task processing
watch 'temporal workflow count --namespace production'

# 3. Verify worker connectivity
./scripts/check_worker_health.sh

# 4. Spot-check critical workflows
temporal workflow describe --workflow-id critical-workflow-123
```

#### Phase 4: Monitor & Stabilize (Target: Ongoing)

```bash
# Set up enhanced monitoring
./scripts/enable_enhanced_monitoring.sh --region=secondary

# Monitor key metrics every 5 minutes for first hour
watch -n 300 '
  echo "=== Workflow Throughput ==="
  temporal workflow count --namespace production

  echo "=== Error Rate ==="
  curl -s http://temporal.example.com:8000/metrics | grep error_rate

  echo "=== Replication Status ==="
  ./scripts/check_replication_status.sh
'
```

### Recovery Checklist

**Immediate (0-5 minutes):**
- [ ] Incident declared in PagerDuty/Opsgenie
- [ ] Incident Commander assigned
- [ ] Primary region failure verified
- [ ] Replication lag checked (<1 minute)
- [ ] Failover initiated to secondary
- [ ] DNS/Load Balancer updated
- [ ] Workflow continuity verified

**Short-term (5-30 minutes):**
- [ ] All critical workflows verified running
- [ ] Worker pools verified healthy
- [ ] Customer communication sent (if customer-facing)
- [ ] Enhanced monitoring enabled
- [ ] Incident log being maintained

**Medium-term (30 minutes - 4 hours):**
- [ ] Root cause analysis initiated
- [ ] Primary region restoration plan created
- [ ] Capacity planning for extended single-region operation
- [ ] Team briefing scheduled

**Long-term (4+ hours):**
- [ ] Primary region restored OR
- [ ] Long-term single-region operation plan approved
- [ ] Incident retrospective scheduled
- [ ] Runbook updates identified

### Rollback Procedure (Restore Primary)

**When to rollback:**
- Primary region is healthy again
- All data is synchronized
- Team approval obtained

```bash
# 1. Verify primary region fully healthy
./scripts/verify_region_health.sh primary --thorough

# 2. Verify data synchronization
./scripts/verify_data_sync.sh --source=secondary --target=primary

# 3. Enable replication from secondary to primary
temporal operator cluster enable-replication \
  --source=secondary \
  --target=primary

# 4. Wait for replication lag <10 seconds
while true; do
  LAG=$(./scripts/check_replication_lag.sh)
  echo "Replication lag: $LAG seconds"
  if [ "$LAG" -lt 10 ]; then
    break
  fi
  sleep 5
done

# 5. Failover back to primary
temporal operator cluster failover \
  --source-cluster=secondary \
  --target-cluster=primary

# 6. Update DNS back to primary
./scripts/update_dns.sh --primary=primary --ttl=60

# 7. Verify everything working
./scripts/comprehensive_health_check.sh
```

### Success Criteria

✅ **Failover successful if:**
- RTO: <5 minutes from detection to recovery
- RPO: <1 minute of replication lag at failover
- No workflow data loss
- All workflows resume processing
- No customer-visible errors post-failover

### Post-Incident

**Immediate:**
- [ ] Document timeline
- [ ] Capture metrics (RTO/RPO achieved)
- [ ] Identify issues encountered

**Within 48 hours:**
- [ ] Incident retrospective meeting
- [ ] Root cause analysis published
- [ ] Runbook improvements identified
- [ ] Action items created and assigned

---

## Scenario 2: Database Failure

### Description
Primary database becomes unavailable, corrupted, or experiencing critical performance issues.

### Detection

**Automated Alerts:**
```promql
# Database down
probe_success{job="postgresql"} == 0

# Database latency high
pg_stat_database_blks_read_time > 1000

# Connection pool exhausted
database_connection_pool_active / database_connection_pool_max > 0.95
```

**Manual Verification:**
```bash
# Check database health
psql -h db-primary.example.com -U temporal -c "SELECT version();"

# Check replication
psql -h db-primary.example.com -U temporal -c "SELECT * FROM pg_stat_replication;"

# Check for corruption
psql -h db-primary.example.com -U temporal -c "SELECT datname FROM pg_database WHERE datistemplate = false;"
```

### Response Procedure

#### Phase 1: Assess Damage (Target: 2 minutes)

```bash
# 1. Determine failure type
./scripts/db_health_check.sh --detailed

# Failure types:
# - Total failure: Database completely unreachable
# - Performance degradation: High latency, not serving requests
# - Corruption: Serving incorrect data
# - Replication failure: Primary up, replicas down

# 2. Check database replicas
./scripts/check_db_replicas.sh

# 3. Determine recovery strategy
# - If replicas healthy: Promote replica
# - If replicas also down: Restore from backup
# - If corruption: Restore from last known good backup
```

#### Phase 2: Database Failover (Target: 3 minutes)

**Option A: Promote Replica (Preferred)**

```bash
# 1. Promote read replica to primary
aws rds promote-read-replica \
  --db-instance-identifier temporal-db-replica-1 \
  --region us-west-2

# 2. Wait for promotion
aws rds wait db-instance-available \
  --db-instance-identifier temporal-db-replica-1

# 3. Update Temporal configuration
./scripts/update_db_endpoint.sh --new-primary=temporal-db-replica-1.example.com

# 4. Restart Temporal services (rolling)
./scripts/rolling_restart.sh --component=all --max-down=1

# 5. Verify database connectivity
./scripts/verify_db_connectivity.sh
```

**Option B: Restore from Backup**

```bash
# 1. Identify latest good backup
aws rds describe-db-snapshots \
  --db-instance-identifier temporal-db-primary \
  --query 'sort_by(DBSnapshots, &SnapshotCreateTime)[-1]'

# 2. Restore from snapshot
aws rds restore-db-instance-from-db-snapshot \
  --db-instance-identifier temporal-db-restored \
  --db-snapshot-identifier temporal-db-snapshot-20251122 \
  --db-instance-class db.r6g.2xlarge

# 3. Wait for restore (15-30 minutes)
aws rds wait db-instance-available \
  --db-instance-identifier temporal-db-restored

# 4. Update Temporal configuration
./scripts/update_db_endpoint.sh --new-primary=temporal-db-restored.example.com

# 5. Restart Temporal services
./scripts/rolling_restart.sh --component=all

# WARNING: RPO = time since last backup (typically 5-30 minutes)
```

#### Phase 3: Verify Recovery

```bash
# 1. Verify database operations
psql -h NEW_DB_HOST -U temporal -c "SELECT COUNT(*) FROM executions;"

# 2. Verify Temporal services healthy
temporal operator cluster health

# 3. Verify workflows processing
temporal workflow list --namespace production | head -20

# 4. Check for errors
tail -f /var/log/temporal/*.log | grep -i error
```

### Recovery Checklist

**Immediate (0-5 minutes):**
- [ ] Database failure type identified
- [ ] Replica status checked
- [ ] Recovery strategy selected
- [ ] Incident declared

**Short-term (5-15 minutes):**
- [ ] Database failover/restore initiated
- [ ] Temporal services configuration updated
- [ ] Services restarted
- [ ] Basic connectivity verified

**Medium-term (15-60 minutes):**
- [ ] Full functionality verified
- [ ] Data integrity checks run
- [ ] Backup replicas provisioned
- [ ] Monitoring restored

### Success Criteria

✅ **Recovery successful if:**
- Database serving requests within 5 minutes
- No data corruption detected
- All Temporal services connected
- Workflows resuming processing
- RPO < 1 minute (replica promotion) or RPO < 30 minutes (backup restore)

---

## Scenario 3: Security Breach

### Description
Suspected or confirmed security incident affecting Temporal Server.

### Detection

**Indicators:**
- High authentication failure rate
- Namespace isolation violations
- Certificate pinning failures
- Unusual API activity
- Alert from security team

**Automated Alerts:**
```promql
# See: config/prometheus-alerts.yml
# - NamespaceIsolationViolation
# - HighAuthenticationFailureRate
# - CertificatePinningFailure
```

### Response Procedure

#### Phase 1: Contain (Target: Immediate)

```bash
# 1. STOP - Do not destroy evidence
# Document everything, take snapshots

# 2. Assess threat
./scripts/security_assess_threat.sh

# 3. Isolate affected systems (if identified)
# DO NOT SHUT DOWN - preserve logs and memory

# 4. Enable enhanced audit logging
./scripts/enable_enhanced_audit.sh

# 5. Notify security team
./scripts/notify_security_team.sh --severity=critical
```

#### Phase 2: Investigate (Target: 15-30 minutes)

```bash
# 1. Collect audit logs
./scripts/collect_audit_logs.sh --last=24h --output=/secure/incident-logs/

# 2. Analyze authentication failures
grep '"audit_event":true' /var/log/temporal/*.log | \
  grep AuthNFailure | jq . > /secure/auth-failures.json

# 3. Check for unauthorized access
grep '"audit_event":true' /var/log/temporal/*.log | \
  grep '"decision":"allow"' | jq . > /secure/allowed-access.json

# 4. Identify compromised accounts
./scripts/identify_compromised_accounts.sh

# 5. Check for data exfiltration
./scripts/check_data_exfiltration.sh
```

#### Phase 3: Remediate (Target: Variable)

**If credentials compromised:**

```bash
# 1. Revoke compromised credentials immediately
./scripts/revoke_credentials.sh --user=compromised-user

# 2. Force JWT key rotation (emergency procedure)
# See: docs/operations/JWT_KEY_ROTATION.md (Emergency section)
./scripts/emergency_jwt_rotation.sh

# 3. Rotate database passwords
# See: docs/operations/SECRETS_ROTATION_RUNBOOK.md
./scripts/rotate_db_passwords.sh --emergency

# 4. Rotate TLS certificates
# See: docs/operations/CERTIFICATE_ROTATION.md (Emergency section)
./scripts/emergency_cert_rotation.sh
```

**If malicious code detected:**

```bash
# 1. Isolate affected workers
./scripts/isolate_workers.sh --workers=suspicious-worker-pool

# 2. Terminate suspicious workflows
temporal workflow terminate --workflow-id=suspicious-wf-id --reason="Security incident"

# 3. Scan for malicious workflows
./scripts/scan_workflows.sh --pattern=malicious-pattern
```

### Recovery Checklist

**Immediate (0-15 minutes):**
- [ ] Incident declared (severity: critical)
- [ ] Security team notified
- [ ] Enhanced audit logging enabled
- [ ] Initial threat assessment complete
- [ ] Containment measures applied

**Short-term (15-60 minutes):**
- [ ] Audit logs collected
- [ ] Compromised accounts identified
- [ ] Credentials revoked
- [ ] Emergency rotation initiated

**Medium-term (1-4 hours):**
- [ ] All compromised credentials rotated
- [ ] Affected systems scanned
- [ ] Data breach assessed
- [ ] Customer communication (if required)

**Long-term (4+ hours):**
- [ ] Forensic analysis complete
- [ ] Security controls enhanced
- [ ] Incident report filed
- [ ] Compliance notifications (if required)

### Success Criteria

✅ **Containment successful if:**
- Unauthorized access stopped
- All compromised credentials rotated
- No further suspicious activity
- Customer data protected

---

## Scenario 4: Cascading Failure

### Description
Multiple components failing in sequence, potentially causing complete system unavailability.

### Detection

**Symptoms:**
- Multiple alerts firing simultaneously
- Services failing in sequence
- Resource exhaustion (CPU, memory, disk)
- Database connection pool exhaustion

### Response Procedure

#### Phase 1: Stop the Cascade (Target: 1 minute)

```bash
# 1. Identify root cause
./scripts/cascade_analysis.sh

# Common root causes:
# - Resource exhaustion
# - Memory leak
# - Database overload
# - Network saturation
# - Infinite retry loop

# 2. Apply circuit breaker
# Stop the source of load immediately

# If bad workflow causing load:
temporal workflow terminate --workflow-id=problematic-workflow --reason="Cascading failure"

# If namespace causing load:
temporal operator namespace update production --rate-limit=100  # Reduce

# If database overloaded:
./scripts/enable_connection_throttling.sh
```

#### Phase 2: Stabilize (Target: 5 minutes)

```bash
# 1. Restart services in correct order
# Order: Database → History → Matching → Frontend → Workers

./scripts/controlled_restart.sh --order=dependency

# 2. Gradually restore capacity
# Start with 10% capacity, increase slowly

./scripts/gradual_capacity_restore.sh --start-percent=10

# 3. Monitor for recurrence
./scripts/monitor_stability.sh --duration=10m
```

### Recovery Checklist

**Immediate:**
- [ ] Circuit breaker applied
- [ ] Root cause identified
- [ ] Load source stopped

**Short-term:**
- [ ] Services restarted in order
- [ ] Capacity restored gradually
- [ ] Stability confirmed

---

## Scenario 5: Network Partition

### Description
Network split causing split-brain scenario between regions or services.

### Response Procedure

#### Phase 1: Detect Split-Brain (Target: 2 minutes)

```bash
# 1. Verify network partition
./scripts/check_network_partition.sh

# 2. Identify active primary
./scripts/identify_active_primary.sh

# 3. Check for dual-primary scenario (CRITICAL)
./scripts/check_split_brain.sh
```

#### Phase 2: Resolve Split-Brain (Target: 3 minutes)

```bash
# 1. If dual-primary detected, force one to standby
# Choose based on: recency of data, number of active workflows

./scripts/force_standby.sh --cluster=secondary --reason="Network partition"

# 2. Verify single primary
./scripts/verify_single_primary.sh

# 3. Re-establish replication
./scripts/restore_replication.sh
```

### Success Criteria

✅ **Resolution successful if:**
- Single primary confirmed
- No data divergence
- Replication restored
- No workflows lost

---

## Testing & Validation

### Quarterly DR Test Schedule

**Q1:** Regional failover test (staging)
**Q2:** Database failure test (staging)
**Q3:** Full DR simulation (staging)
**Q4:** Multi-failure scenario (staging)

### Test Procedure

```bash
# 1. Schedule with team (2 weeks notice)
./scripts/schedule_dr_test.sh --date=2026-01-15 --scenario=region-failover

# 2. Execute test in staging
./scripts/execute_dr_test.sh --environment=staging --scenario=region-failover

# 3. Measure RTO/RPO
./scripts/measure_dr_metrics.sh

# 4. Document results
./scripts/generate_dr_report.sh --output=dr-test-2026-01-15.md

# 5. Update runbook based on learnings
```

### Success Metrics

- **RTO Achieved:** <5 minutes (target)
- **RPO Achieved:** <1 minute (target)
- **Data Loss:** 0 workflows
- **Test Frequency:** Quarterly
- **Runbook Accuracy:** >95%

---

## Appendix

### A. Emergency Scripts Reference

All emergency scripts located in: `/scripts/dr/`

- `verify_region_health.sh` - Check region status
- `initiate_failover.sh` - Start failover process
- `update_dns.sh` - Update DNS records
- `check_replication_lag.sh` - Measure replication lag
- `emergency_jwt_rotation.sh` - Emergency JWT key rotation
- `emergency_cert_rotation.sh` - Emergency certificate rotation

### B. Monitoring Dashboards

- **Primary Dashboard:** https://grafana.example.com/d/temporal-overview
- **Security Dashboard:** https://grafana.example.com/d/temporal-security
- **DR Dashboard:** https://grafana.example.com/d/temporal-dr

### C. Contact Information

**Escalation Path:**
1. On-Call SRE (PagerDuty)
2. Platform Lead
3. Engineering Director
4. CTO

**External Contacts:**
- Cloud Provider Support: [support number]
- Database Support: [support number]
- Security Vendor: [support number]

---

**Document Maintenance:**
- Review: Quarterly
- Update: After each incident or DR test
- Owner: SRE Team Lead
- Approver: VP Engineering

**Version History:**
- 1.0 (2025-11-22): Initial version
- Next review: 2026-02-22

**Reference:**
- PROJECT_IMPLEMENTATION_ROADMAP.md (Sprint 3)
- SECURITY_OPERATOR_GUIDE.md
- docs/operations/SECRETS_ROTATION_RUNBOOK.md
