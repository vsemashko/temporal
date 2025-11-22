# Deployment Verification Checklist

**Purpose:** Comprehensive checklist for verifying security feature deployment
**Reference:** PROJECT_IMPLEMENTATION_ROADMAP.md (Weeks 2-4)
**Version:** 1.0
**Last Updated:** 2025-11-22

---

## Pre-Deployment (Week 1)

### Configuration Review

- [ ] **Staging configuration reviewed**
  - File: `config/staging-security-enabled.yaml`
  - All security features enabled
  - TLS certificates generated and valid
  - Database credentials configured (use secrets manager)
  - JWT key provider configured

- [ ] **Prometheus alerting configured**
  - File: `config/prometheus-alerts.yml`
  - All 35 alerting rules loaded
  - Alert routing configured
  - PagerDuty/Opsgenie integration tested
  - Test alert successfully received

- [ ] **Monitoring dashboards created**
  - Security dashboard in Grafana
  - Operational dashboard updated
  - DR dashboard created
  - All 12 security metrics visible

- [ ] **TLS Certificates prepared**
  - [ ] CA certificate generated
  - [ ] Server certificates generated (frontend, history, matching, worker)
  - [ ] Client certificates generated (for mTLS)
  - [ ] Certificate expiration >90 days
  - [ ] Certificates uploaded to secure storage
  - [ ] Certificate fingerprints calculated for pinning

- [ ] **Secrets management**
  - [ ] Database passwords in secrets manager
  - [ ] JWT signing keys secured
  - [ ] TLS private keys encrypted at rest
  - [ ] Secrets rotation schedule documented

- [ ] **Rollback plan prepared**
  - [ ] Previous configuration backed up
  - [ ] Rollback procedure documented
  - [ ] Rollback tested in dev environment
  - [ ] Rollback owner assigned

---

## Week 2: Staging Deployment

### Deployment (Day 8-9)

- [ ] **Services deployed to staging**
  ```bash
  # Verification command:
  kubectl get pods -n temporal-staging
  # Expected: All pods Running
  ```

- [ ] **All services started successfully**
  - [ ] Frontend service: Port 7233 (gRPC), 7243 (HTTP)
  - [ ] History service: Port 7234
  - [ ] Matching service: Port 7235
  - [ ] Worker service: Port 7239
  - [ ] Metrics endpoint: Port 8000

  ```bash
  # Verification commands:
  curl -f http://frontend:7243/health || echo "FAIL"
  curl -f http://frontend:8000/metrics | grep "up 1" || echo "FAIL"
  ```

- [ ] **Configuration validated**
  ```bash
  # Check configuration loaded correctly:
  temporal server config validate --config=/etc/temporal/config/staging-security-enabled.yaml

  # Verify TLS enabled:
  openssl s_client -connect frontend:7233 </dev/null 2>/dev/null | grep "Cipher"
  # Expected: TLS 1.3, AES_128_GCM_SHA256 or similar

  # Verify no passwords in logs:
  grep -r "password.*temporal" /var/log/temporal/ && echo "FAIL: Password leaked!" || echo "PASS"
  ```

### Security Feature Verification (Day 10-11)

#### 1. Audit Logging

- [ ] **Audit logging enabled**
  ```bash
  # Test: Trigger authorization event
  temporal workflow start --task-queue test --type TestWorkflow --namespace production

  # Verify audit event captured:
  tail -100 /var/log/temporal/temporal.log | grep '"audit_event":true' | jq .

  # Expected fields:
  # - timestamp
  # - event_type: "AuthZSuccess"
  # - user_id: <from JWT>
  # - source_ip: <client IP>
  # - namespace: "production"
  # - api_name: "StartWorkflowExecution"
  # - decision: "allow"
  ```

- [ ] **All event types captured**
  - [ ] AuthZSuccess events
  - [ ] AuthZFailure events
  - [ ] AuthNSuccess events
  - [ ] AuthNFailure events (test with invalid JWT)

#### 2. Authentication Rate Limiting

- [ ] **Rate limiting enabled**
  ```bash
  # Check configuration:
  grep -A5 "rateLimit:" /etc/temporal/config/staging-security-enabled.yaml

  # Expected:
  # enabled: true
  # maxFailuresPerMinute: 10
  # lockoutDuration: 5m
  ```

- [ ] **Rate limiting working**
  ```bash
  # Simulate brute force (15 failed auth attempts):
  for i in {1..15}; do
    temporal workflow start --task-queue test --type Test --identity invalid-$i 2>&1 | grep -q "authentication failed"
  done

  # Verify lockout after 10th attempt:
  temporal workflow start --task-queue test --type Test --identity invalid-16 2>&1 | grep "rate limit exceeded"
  # Expected: Error indicating rate limit

  # Check metrics:
  curl -s http://frontend:8000/metrics | grep auth_rate_limited_total
  # Expected: auth_rate_limited_total >0
  ```

- [ ] **Lockout expires correctly**
  ```bash
  # Wait 5 minutes, then retry:
  sleep 300
  temporal workflow start --task-queue test --type Test --identity valid-user
  # Expected: Success (rate limit reset)
  ```

#### 3. Context Timeout Enforcement

- [ ] **Timeout enforcement enabled**
  ```bash
  # Check configuration:
  grep -A10 "timeout:" /etc/temporal/config/staging-security-enabled.yaml

  # Expected:
  # enabled: true
  # defaultTimeout: 60s
  ```

- [ ] **Timeouts enforced**
  ```bash
  # Test with slow operation (if available):
  # Monitor metrics:
  curl -s http://frontend:8000/metrics | grep service_request_timeout

  # Expected metrics:
  # - service_request_timeout_enforced (counter increasing)
  # - service_request_timeout_exceeded (if any timeouts occur)
  ```

#### 4. Configuration Sanitization

- [ ] **Passwords not logged**
  ```bash
  # Trigger configuration dump (e.g., startup logs):
  kubectl logs -n temporal-staging frontend-pod | grep -i "password\|secret\|key"

  # Expected: All sensitive fields show "REDACTED" or similar
  # Fail if actual passwords visible
  ```

- [ ] **TLS keys not logged**
  ```bash
  # Check for keyData in logs:
  kubectl logs -n temporal-staging frontend-pod | grep "keyData"

  # Expected: Either not present or showing "REDACTED"
  ```

#### 5. Certificate Pinning (if remote clusters configured)

- [ ] **Certificate pinning configured**
  ```bash
  # Check configuration:
  grep -A10 "pinnedCertificates:" /etc/temporal/config/staging-security-enabled.yaml

  # Expected:
  # enabled: true
  # strictPinning: true
  # fingerprints: [...]
  ```

- [ ] **Pinning validation working**
  ```bash
  # Check metrics:
  curl -s http://frontend:8000/metrics | grep cert_pin_validation

  # Expected:
  # - cert_pin_validation_success >0 (if remote connections active)
  # - cert_pin_validation_failure == 0 (no failures)
  ```

- [ ] **Pin mismatch detected** (test in non-strict mode)
  ```bash
  # Test with incorrect fingerprint (in test env):
  # Update config with wrong fingerprint, restart, verify failure metric increases
  ```

#### 6. TLS Configuration

- [ ] **TLS 1.3 enforced**
  ```bash
  # Test TLS version:
  openssl s_client -connect frontend:7233 -tls1_2 </dev/null 2>&1 | grep "Cipher"
  # Expected: Connection refused or downgrade to TLS 1.2 rejected

  openssl s_client -connect frontend:7233 -tls1_3 </dev/null 2>&1 | grep "Cipher"
  # Expected: Connection successful with TLS 1.3
  ```

- [ ] **Cipher suites correct**
  ```bash
  # Verify FIPS-compliant ciphers:
  openssl s_client -connect frontend:7233 </dev/null 2>&1 | grep "Cipher"

  # Expected: TLS_AES_128_GCM_SHA256, TLS_AES_256_GCM_SHA384, or TLS_CHACHA20_POLY1305_SHA256
  ```

- [ ] **mTLS working (if enabled)**
  ```bash
  # Test without client certificate:
  openssl s_client -connect frontend:7233 </dev/null 2>&1 | grep "Verify return code"
  # Expected: Verification error (client cert required)

  # Test with valid client certificate:
  openssl s_client -connect frontend:7233 \
    -cert /path/to/client.pem \
    -key /path/to/client-key.pem \
    </dev/null 2>&1 | grep "Verify return code"
  # Expected: Verify return code: 0 (ok)
  ```

### Performance Validation (Day 12-14)

#### Load Testing

- [ ] **Load test executed**
  ```bash
  # Run load test (1000 req/sec for 10 minutes):
  ./scripts/load_test.sh --rps=1000 --duration=10m --endpoint=staging

  # Or using hey/wrk:
  hey -z 10m -q 1000 -m POST \
    -H "Authorization: Bearer $JWT_TOKEN" \
    http://frontend:7243/api/v1/namespaces/default/workflows
  ```

- [ ] **Latency acceptable**
  ```bash
  # Check P50, P95, P99 latency:
  curl -s http://frontend:8000/metrics | grep http_request_duration_seconds

  # Expected P95 latency: <50ms (target: <40ms)
  # Expected overhead from security features: <2ms
  ```

- [ ] **Throughput acceptable**
  ```bash
  # Check request rate:
  curl -s http://frontend:8000/metrics | grep http_requests_total

  # Expected: >1000 req/sec sustained
  ```

- [ ] **Error rate acceptable**
  ```bash
  # Check error rate:
  ERROR_RATE=$(curl -s http://frontend:8000/metrics | \
    awk '/http_requests_total.*status="5/ {errors+=$2} END {print errors}')

  TOTAL_REQUESTS=$(curl -s http://frontend:8000/metrics | \
    awk '/http_requests_total/ {total+=$2} END {print total}')

  # Expected: Error rate <0.1%
  echo "scale=4; $ERROR_RATE / $TOTAL_REQUESTS * 100" | bc
  ```

- [ ] **Resource consumption acceptable**
  ```bash
  # Check memory usage:
  kubectl top pods -n temporal-staging

  # Expected overhead: <100MB per pod
  # Expected CPU overhead: <1% per pod
  ```

#### Audit Log Volume

- [ ] **Audit log volume measured**
  ```bash
  # Count audit events per second:
  grep '"audit_event":true' /var/log/temporal/temporal.log | \
    awk -F'"timestamp":"' '{print $2}' | awk -F'"' '{print $1}' | \
    cut -d: -f1-2 | uniq -c

  # Verify log rotation configured for expected volume
  ```

- [ ] **Log retention configured**
  ```bash
  # Check logrotate config:
  cat /etc/logrotate.d/temporal

  # Expected:
  # - Daily rotation
  # - Compress old logs
  # - Retain 30 days
  # - Max size 1GB before rotation
  ```

### Monitoring Verification

- [ ] **All 12 security metrics collecting**
  ```bash
  # Check metrics endpoint:
  curl -s http://frontend:8000/metrics | grep -E "service_authentication|service_authorization|service_rate_limit|service_tls|service_jwt|service_namespace_isolation|service_config_validation|service_audit_log|service_cert_expiry|service_secret_rotation|service_mTLS|service_security_scanner"

  # Expected: All metrics present with values
  ```

- [ ] **Prometheus scraping successfully**
  ```bash
  # Check Prometheus targets:
  curl -s http://prometheus:9090/api/v1/targets | jq '.data.activeTargets[] | select(.labels.job=="temporal")'

  # Expected: All targets "up"
  ```

- [ ] **Alerts firing correctly (test)**
  ```bash
  # Trigger test alert:
  # 1. Simulate high auth failures
  # 2. Verify alert fires in Alertmanager
  # 3. Verify notification received (PagerDuty/Slack)

  # Check Alertmanager:
  curl -s http://alertmanager:9093/api/v2/alerts | jq .
  ```

- [ ] **Grafana dashboards populated**
  - [ ] Security dashboard showing all 12 metrics
  - [ ] No "No Data" panels
  - [ ] All graphs rendering correctly
  - [ ] Alerts visible in dashboard

---

## Week 3: Production Canary (10%)

### Canary Deployment (Day 15-18)

- [ ] **Canary instances deployed (10%)**
  ```bash
  # Verify 10% of traffic routing to canary:
  kubectl get pods -l version=canary -n temporal-production

  # Expected: ~10% of total pod count
  ```

- [ ] **Traffic routing verified**
  ```bash
  # Check load balancer weights:
  # (specific command depends on LB solution)

  # Verify traffic distribution:
  ./scripts/check_traffic_distribution.sh

  # Expected: ~10% to canary, ~90% to stable
  ```

### Hourly Monitoring (First 24 hours)

- [ ] **Hour 1: Stability check**
  - [ ] Error rate <0.1%: ______ (actual: _____%)
  - [ ] P95 latency <50ms: ______ (actual: _____ms)
  - [ ] No critical alerts: ______
  - [ ] Workflows processing normally: ______

- [ ] **Hour 2: Stability check**
  - [ ] Error rate <0.1%: ______ (actual: _____%)
  - [ ] P95 latency <50ms: ______ (actual: _____ms)
  - [ ] No critical alerts: ______

- [ ] **Hour 4: Stability check**
  - [ ] Error rate <0.1%: ______ (actual: _____%)
  - [ ] P95 latency <50ms: ______ (actual: _____ms)

- [ ] **Hour 8: Stability check**
  - [ ] Error rate <0.1%: ______ (actual: _____%)
  - [ ] P95 latency <50ms: ______ (actual: _____ms)

- [ ] **Hour 12: Stability check**
  - [ ] Error rate <0.1%: ______ (actual: _____%)
  - [ ] P95 latency <50ms: ______ (actual: _____ms)

- [ ] **Hour 24: Full day assessment**
  - [ ] Error rate <0.1% for 24h: ______
  - [ ] P95 latency <50ms for 24h: ______
  - [ ] No critical incidents: ______
  - [ ] Team consensus: ______ (proceed/rollback)

### Canary Success Criteria (Day 19-21)

- [ ] **72-hour soak test passed**
  - [ ] Canary stable for 72 hours
  - [ ] No performance degradation
  - [ ] No increase in errors
  - [ ] No customer complaints
  - [ ] No critical alerts

**Decision Point:** Proceed to full rollout: ☐ Yes ☐ No ☐ Rollback

---

## Week 4: Full Production Rollout

### Gradual Rollout Schedule

- [ ] **Day 22: 20% rollout**
  - Deployed at: __________ (timestamp)
  - Error rate: ______%
  - P95 latency: ______ms
  - Issues encountered: _____________
  - Status: ☐ Proceed ☐ Pause ☐ Rollback

- [ ] **Day 23: 40% rollout**
  - Deployed at: __________
  - Error rate: ______%
  - P95 latency: ______ms
  - Issues encountered: _____________
  - Status: ☐ Proceed ☐ Pause ☐ Rollback

- [ ] **Day 24: 60% rollout**
  - Deployed at: __________
  - Error rate: ______%
  - P95 latency: ______ms
  - Issues encountered: _____________
  - Status: ☐ Proceed ☐ Pause ☐ Rollback

- [ ] **Day 25: 80% rollout**
  - Deployed at: __________
  - Error rate: ______%
  - P95 latency: ______ms
  - Issues encountered: _____________
  - Status: ☐ Proceed ☐ Pause ☐ Rollback

- [ ] **Day 26: 100% rollout**
  - Deployed at: __________
  - Error rate: ______%
  - P95 latency: ______ms
  - **DEPLOYMENT COMPLETE**: ☐ Yes

### Post-Deployment (Day 27-28)

#### Feature Enablement

- [ ] **Certificate pinning enabled for remote clusters**
  ```bash
  # Generate fingerprints for all remote clusters:
  for cluster in cluster1 cluster2 cluster3; do
    echo "Cluster: $cluster"
    openssl x509 -in /path/to/$cluster.crt -noout -fingerprint -sha256
  done

  # Update configuration with fingerprints
  # Apply configuration
  # Verify pinning working (check metrics)
  ```

- [ ] **SIEM integration configured**
  ```bash
  # Configure log shipping to SIEM:
  # - Splunk: Configure forwarder
  # - Elasticsearch: Configure Filebeat
  # - Datadog: Configure agent

  # Verify audit events appearing in SIEM:
  # Search for: audit_event:true

  # Create SIEM dashboards and alerts
  ```

#### First Secrets Rotation (Day 29-30)

- [ ] **JWT key rotation executed**
  ```bash
  # Follow: docs/operations/JWT_KEY_ROTATION.md

  # Phase 1: Add new key (dual-key period)
  # Phase 2: Update IdP to issue tokens with new key
  # Phase 3: Wait for old tokens to expire (24-48h)
  # Phase 4: Remove old key

  # Verification:
  # - All new tokens signed with new key
  # - Old tokens still validated during dual-key period
  # - No authentication failures
  ```

- [ ] **Rotation verified**
  - [ ] New JWT key active
  - [ ] Old key removed (after expiration period)
  - [ ] No authentication errors
  - [ ] Next rotation scheduled: __________

---

## Post-Deployment Verification

### Final Checks (Week 5)

- [ ] **All security features operational in production**
  - [ ] Audit logging: Capturing all events
  - [ ] Rate limiting: Protecting against brute force
  - [ ] Timeout enforcement: Preventing resource exhaustion
  - [ ] Configuration sanitization: No credential leaks
  - [ ] Certificate pinning: Validating remote clusters
  - [ ] TLS 1.3: All connections encrypted
  - [ ] mTLS: Client authentication working

- [ ] **Monitoring fully operational**
  - [ ] All 35 alerts configured
  - [ ] PagerDuty integration working
  - [ ] Grafana dashboards updated
  - [ ] On-call runbooks updated
  - [ ] Team trained on new alerts

- [ ] **Documentation updated**
  - [ ] SECURITY_OPERATOR_GUIDE.md reflects production config
  - [ ] Runbooks updated with production specifics
  - [ ] Team wiki updated
  - [ ] Customer-facing docs updated (if needed)

- [ ] **Compliance requirements met**
  - [ ] Audit logging enabled (PCI-DSS 10.2, HIPAA §164.312(b), SOC 2 CC6.6)
  - [ ] Encryption in transit (TLS 1.3)
  - [ ] Secrets rotation schedule documented
  - [ ] Access controls verified (RBAC, namespace isolation)

### Success Criteria - Final Assessment

✅ **Deployment successful if ALL criteria met:**

**Performance:**
- [ ] P95 latency <50ms: ______ (actual: _____ms)
- [ ] P99 latency <100ms: ______ (actual: _____ms)
- [ ] Error rate <0.1%: ______ (actual: _____%)
- [ ] Throughput >1000 req/sec: ______ (actual: _____ req/sec)

**Security:**
- [ ] All 25 security enhancements operational
- [ ] Zero critical security alerts
- [ ] Audit logging capturing >95% of events
- [ ] Secrets rotation schedule active

**Operational:**
- [ ] Zero critical incidents during rollout
- [ ] Rollback plan tested and ready
- [ ] Team trained on new features
- [ ] Monitoring dashboards operational

**Business:**
- [ ] Zero customer complaints related to deployment
- [ ] No customer-visible downtime
- [ ] SLAs maintained (99.99% uptime)

---

## Rollback Criteria

**ROLLBACK IMMEDIATELY if:**
- [ ] Error rate >1% for >5 minutes
- [ ] P95 latency >100ms for >10 minutes
- [ ] Any critical security alert fires
- [ ] Data corruption detected
- [ ] Customer-visible production outage
- [ ] Memory leak detected (>20% increase in 1 hour)
- [ ] CPU usage >90% sustained for >10 minutes

**Rollback Procedure:**
```bash
# 1. Halt deployment
kubectl rollout pause deployment/temporal-frontend -n temporal-production

# 2. Revert to previous version
kubectl rollout undo deployment/temporal-frontend -n temporal-production
kubectl rollout undo deployment/temporal-history -n temporal-production
kubectl rollout undo deployment/temporal-matching -n temporal-production
kubectl rollout undo deployment/temporal-worker -n temporal-production

# 3. Verify rollback successful
kubectl rollout status deployment/temporal-frontend -n temporal-production

# 4. Monitor for stabilization
./scripts/monitor_post_rollback.sh --duration=30m

# 5. Incident review
./scripts/create_incident_report.sh --type=deployment-rollback
```

---

## Sign-off

**Staging Deployment (Week 2):**
- Deployed by: _________________ Date: _______
- Verified by: _________________ Date: _______
- Approved by: _________________ Date: _______

**Production Canary (Week 3):**
- Deployed by: _________________ Date: _______
- Verified by: _________________ Date: _______
- Approved by: _________________ Date: _______

**Production Full Rollout (Week 4):**
- Deployed by: _________________ Date: _______
- Verified by: _________________ Date: _______
- Approved by: _________________ Date: _______

**Final Sign-off:**
- Engineering Lead: _________________ Date: _______
- Security Lead: _________________ Date: _______
- VP Engineering: _________________ Date: _______

---

**Deployment Complete:** ☐
**Date Completed:** __________
**Final Security Score:** ______ / 10 (target: 9.8+)

**Reference:**
- PROJECT_IMPLEMENTATION_ROADMAP.md
- SECURITY_OPERATOR_GUIDE.md
- SECURITY_CHECKLIST.md
