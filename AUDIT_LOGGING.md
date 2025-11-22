# Audit Logging Guide

**Last Updated:** 2025-11-22
**Phase:** 3.2 - Comprehensive Audit Logging

---

## Overview

Temporal Server provides comprehensive audit logging for all authorization decisions and security-relevant events. Audit logs enable:
- Compliance (PCI-DSS, HIPAA, SOC 2)
- Security incident investigation
- Forensic analysis
- Anomaly detection
- Access pattern analysis

---

## Quick Start

### Enable Audit Logging

Audit logging is enabled by default and logs all authorization decisions. Logs are written to the standard Temporal logger with the `audit_event=true` tag.

### Query Audit Logs

Audit logs are structured JSON, making them easy to query:

```bash
# Find all denied requests
grep '"decision":"deny"' temporal.log | jq .

# Find all requests by a specific user
grep '"user_id":"user@example.com"' temporal.log | jq .

# Find all admin operations
grep '"namespace_role":"admin"' temporal.log | jq .

# Count authorization failures by user
grep '"event_type":"authorization.failure"' temporal.log | \
  jq -r '.user_id' | sort | uniq -c | sort -rn
```

---

## Audit Event Structure

All audit events follow this JSON structure:

```json
{
  "timestamp": "2025-11-22T10:30:45.123Z",
  "event_type": "authorization.success",
  "user_id": "user@example.com",
  "source_ip": "192.168.1.100:54321",
  "namespace": "production",
  "api_name": "/temporal.api.workflowservice.v1.WorkflowService/StartWorkflowExecution",
  "decision": "allow",
  "reason": "User has namespace:write permission",
  "system_role": "undefined",
  "namespace_role": "write",
  "metadata": {
    "subject": "user@example.com",
    "groups": ["engineers", "on-call"]
  }
}
```

### Event Types

| Event Type | Description |
|------------|-------------|
| `authorization.success` | Authorization decision was allow |
| `authorization.failure` | Authorization decision was deny |
| `authentication.success` | Authentication succeeded |
| `authentication.failure` | Authentication failed |
| `config.change` | Configuration was changed |
| `certificate.rotation` | Certificate was rotated |

### Decision Values

- `allow` - Request was authorized
- `deny` - Request was denied

---

## Common Queries

### Security Investigations

**Find unauthorized access attempts:**
```bash
grep '"decision":"deny"' temporal.log | \
  jq -r '[.timestamp, .user_id, .source_ip, .api_name, .reason] | @tsv'
```

**Track access to specific namespace:**
```bash
grep '"namespace":"sensitive-namespace"' temporal.log | jq .
```

**Find privilege escalation attempts:**
```bash
grep '"api_name":"/temporal.api.operatorservice' temporal.log | \
  grep '"decision":"deny"' | jq .
```

### Compliance Auditing

**List all admin operations:**
```bash
grep '"namespace_role":"admin"' temporal.log | \
  jq -r '[.timestamp, .user_id, .api_name, .namespace] | @tsv'
```

**Generate access report for audit:**
```bash
# All access by date range
awk '/2025-11-22T10:00/,/2025-11-22T11:00/' temporal.log | \
  grep audit_event=true | jq .
```

### Performance Analysis

**Count requests by user:**
```bash
grep audit_event=true temporal.log | \
  jq -r '.user_id' | sort | uniq -c | sort -rn | head -20
```

**Most denied APIs:**
```bash
grep '"decision":"deny"' temporal.log | \
  jq -r '.api_name' | sort | uniq -c | sort -rn | head -10
```

---

## SIEM Integration

### Splunk

Forward audit logs to Splunk:

```bash
# Use Splunk Universal Forwarder
# /opt/splunkforwarder/etc/system/local/inputs.conf
[monitor:///var/log/temporal/temporal.log]
disabled = false
index = temporal_audit
sourcetype = temporal:audit
```

**Splunk Query:**
```spl
index=temporal_audit audit_event=true
| spath input=event_json
| table timestamp, event_type, user_id, source_ip, namespace, decision, reason
| where decision="deny"
```

### Elasticsearch

Ingest audit logs:

```json
{
  "mappings": {
    "properties": {
      "timestamp": { "type": "date" },
      "event_type": { "type": "keyword" },
      "user_id": { "type": "keyword" },
      "source_ip": { "type": "ip" },
      "namespace": { "type": "keyword" },
      "api_name": { "type": "keyword" },
      "decision": { "type": "keyword" },
      "reason": { "type": "text" }
    }
  }
}
```

**Query denied requests:**
```json
{
  "query": {
    "bool": {
      "must": [
        { "term": { "decision": "deny" }},
        { "range": { "timestamp": { "gte": "now-1h" }}}
      ]
    }
  }
}
```

---

## Alerting

### Example Alert Rules

**Critical: Multiple Failed Authorization Attempts**
```yaml
# Prometheus Alert
- alert: MultipleAuthorizationFailures
  expr: rate(temporal_authorization_denied_total[5m]) > 10
  for: 5m
  annotations:
    summary: "High rate of authorization failures"
    description: "{{ $value }} authorization failures per second"
```

**High: Admin Access from Unusual IP**
```bash
# Script to detect anomalies
#!/bin/bash
# Get list of known admin IPs
KNOWN_IPS="192.168.1.10 192.168.1.11"

# Find admin access from unknown IPs
grep '"namespace_role":"admin"' temporal.log | \
  jq -r '.source_ip' | sort -u | \
  while read ip; do
    if ! echo "$KNOWN_IPS" | grep -q "$ip"; then
      echo "ALERT: Admin access from unknown IP: $ip"
    fi
  done
```

---

## Best Practices

### For Security Teams

1. **Monitor continuously**
   - Set up real-time alerts for denied requests
   - Review audit logs daily
   - Investigate anomalies immediately

2. **Retain logs appropriately**
   - Keep audit logs for compliance period (typically 1-7 years)
   - Use immutable storage (WORM)
   - Implement log rotation

3. **Regular audits**
   - Weekly: Review denied access attempts
   - Monthly: Analyze access patterns
   - Quarterly: Full compliance audit

### For Operators

1. **Log rotation**
   ```bash
   # /etc/logrotate.d/temporal
   /var/log/temporal/temporal.log {
       daily
       rotate 90
       compress
       delaycompress
       missingok
       notifempty
       create 0640 temporal temporal
       sharedscripts
   }
   ```

2. **Monitoring**
   - Track log volume
   - Alert on missing logs
   - Monitor disk space

3. **Backup**
   - Backup audit logs separately
   - Test restore procedures
   - Encrypt backups

---

## Performance Impact

Audit logging has minimal performance impact:
- **CPU**: < 1% overhead (JSON serialization)
- **Memory**: Negligible
- **Disk I/O**: Depends on request rate
- **Latency**: < 0.1ms per request

Audit events are logged asynchronously and do not block request processing.

---

## Troubleshooting

### No Audit Logs Appearing

1. **Check log level**: Audit events log at INFO level
   ```yaml
   log:
     level: info  # Must be info or debug
   ```

2. **Check logger configuration**: Ensure logger is writing to file

3. **Search for audit marker**:
   ```bash
   grep audit_event=true temporal.log
   ```

### Too Many Audit Logs

1. **Filter by severity**: Only log denied requests
2. **Sample**: Log 1% of successful requests
3. **Exclude health checks**: Already excluded by default

### Missing Fields in Audit Events

Some fields are optional:
- `user_id`: Empty if request has no authentication
- `namespace`: Empty for cluster-level operations
- `source_ip`: Empty if peer info unavailable

---

## Advanced Usage

### Custom Audit Queries

**Find concurrent sessions:**
```bash
grep audit_event=true temporal.log | \
  jq -r '[.timestamp, .user_id, .source_ip] | @tsv' | \
  sort -k2,2 -k1,1 | \
  awk '{if ($2 == prev_user && $3 != prev_ip) print "Concurrent:", $0; prev_user=$2; prev_ip=$3}'
```

**Calculate authorization success rate:**
```bash
total=$(grep audit_event=true temporal.log | wc -l)
denied=$(grep '"decision":"deny"' temporal.log | wc -l)
echo "Success rate: $(( 100 - (denied * 100 / total) ))%"
```

---

## Additional Resources

- **Compliance Standards**:
  - PCI-DSS Section 10: Audit Logging Requirements
  - HIPAA §164.312(b): Audit Controls
  - SOC 2: Monitoring Activities

- **SIEM Tools**:
  - Splunk: https://www.splunk.com
  - Elasticsearch: https://www.elastic.co
  - Datadog: https://www.datadoghq.com

- **Internal Documentation**:
  - SECURITY_OPERATOR_GUIDE.md
  - SECURITY_ROADMAP.md

---

## Version History

| Date | Version | Changes |
|------|---------|---------|
| 2025-11-22 | 1.0 | Initial implementation (Phase 3.2) |

---

**Maintained By:** Security Team
**Questions:** security@temporal.io
