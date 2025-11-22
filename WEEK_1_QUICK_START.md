# Week 1 Quick Start Guide

**Purpose:** Get started immediately with security deployment preparation
**Duration:** 7 days (Day 1-7)
**Team:** Security, DevOps, SRE
**Reference:** PROJECT_IMPLEMENTATION_ROADMAP.md (Week 1)

---

## 🎯 Week 1 Objectives

By end of Week 1, you will have:
✅ Team aligned on roadmap and timeline
✅ Staging environment prepared for deployment
✅ Prometheus alerting configured and tested
✅ Grafana security dashboard created
✅ Rollback procedures documented and tested
✅ Team trained on new security features

---

## 📅 Day-by-Day Plan

### Day 1: Team Kickoff & Planning

#### Morning (2 hours)
**📋 Team Meeting: Roadmap Review**

```bash
# Preparation
1. Send calendar invite to all stakeholders:
   - Engineering Lead
   - Security Lead
   - SRE Team
   - Platform Team
   - Product Manager (optional)

2. Meeting agenda:
   - Review PROJECT_IMPLEMENTATION_ROADMAP.md (30 min)
   - Assign owners for each phase (15 min)
   - Discuss timeline and concerns (30 min)
   - Q&A (15 min)
   - Action items and next steps (30 min)

# Expected outcomes:
✅ Everyone understands the 30-day deployment plan
✅ Owners assigned for each week
✅ Deployment windows scheduled
✅ Questions and concerns addressed
```

**Checklist:**
- [ ] Meeting scheduled
- [ ] PROJECT_IMPLEMENTATION_ROADMAP.md reviewed by all attendees
- [ ] Owners assigned:
  - Week 1 (Preparation): __________
  - Week 2 (Staging): __________
  - Week 3 (Canary): __________
  - Week 4 (Full Rollout): __________
- [ ] Deployment windows reserved:
  - Staging: __________ (Week 2, Day 8-9)
  - Canary: __________ (Week 3, Day 17-18)
  - Full rollout: __________ (Week 4, Days 22-26)

#### Afternoon (4 hours)
**🔐 Security Configuration Audit**

```bash
# 1. Verify current security state
cd /home/user/temporal

# 2. Review staging configuration
cat config/staging-security-enabled.yaml

# 3. Identify gaps
grep -A5 "tls:" config/staging-security-enabled.yaml
grep -A5 "authorization:" config/staging-security-enabled.yaml

# 4. Document required certificates and secrets
```

**Checklist:**
- [ ] Current configuration documented
- [ ] TLS certificate requirements identified
- [ ] JWT key provider requirements identified
- [ ] Database secrets requirements identified
- [ ] Gap analysis completed

---

### Day 2: Environment Preparation

#### Morning (3 hours)
**🔑 TLS Certificate Generation**

```bash
# Create certificate directory
mkdir -p /tmp/temporal-certs
cd /tmp/temporal-certs

# 1. Generate CA certificate
openssl genrsa -out ca-key.pem 4096

openssl req -new -x509 -days 365 -key ca-key.pem -out ca.pem \
  -subj "/C=US/ST=CA/L=Seattle/O=Temporal/CN=Temporal CA"

# 2. Generate server certificate (frontend)
openssl genrsa -out frontend-server-key.pem 4096

openssl req -new -key frontend-server-key.pem -out frontend-server.csr \
  -subj "/C=US/ST=CA/L=Seattle/O=Temporal/CN=temporal-frontend.staging.example.com"

openssl x509 -req -days 365 -in frontend-server.csr \
  -CA ca.pem -CAkey ca-key.pem -CAcreateserial \
  -out frontend-server.pem

# 3. Generate client certificate (for mTLS)
openssl genrsa -out client-key.pem 4096

openssl req -new -key client-key.pem -out client.csr \
  -subj "/C=US/ST=CA/L=Seattle/O=Temporal/CN=temporal-client"

openssl x509 -req -days 365 -in client.csr \
  -CA ca.pem -CAkey ca-key.pem -CAcreateserial \
  -out client.pem

# 4. Verify certificates
openssl verify -CAfile ca.pem frontend-server.pem
openssl verify -CAfile ca.pem client.pem

# 5. Check expiration
openssl x509 -in frontend-server.pem -noout -enddate
# Expected: notAfter=[365 days from today]

# 6. Calculate SHA-256 fingerprints (for certificate pinning)
openssl x509 -in frontend-server.pem -noout -fingerprint -sha256
```

**Checklist:**
- [ ] CA certificate generated and verified
- [ ] Server certificates generated for:
  - [ ] Frontend service
  - [ ] History service
  - [ ] Matching service
  - [ ] Worker service
- [ ] Client certificates generated (for mTLS)
- [ ] All certificates verified
- [ ] Certificate expiration > 90 days
- [ ] SHA-256 fingerprints calculated
- [ ] Certificates uploaded to secure storage (e.g., AWS Secrets Manager)

#### Afternoon (3 hours)
**🗄️ Database & Secrets Configuration**

```bash
# 1. Create secrets in secrets manager (example: AWS Secrets Manager)

# Database password
aws secretsmanager create-secret \
  --name temporal/staging/db-password \
  --secret-string "$(openssl rand -base64 32)" \
  --region us-west-2

# JWT signing key (for future use)
aws secretsmanager create-secret \
  --name temporal/staging/jwt-signing-key \
  --secret-string "$(openssl rand -base64 64)" \
  --region us-west-2

# 2. Upload TLS certificates
aws secretsmanager create-secret \
  --name temporal/staging/tls-ca \
  --secret-binary fileb:///tmp/temporal-certs/ca.pem \
  --region us-west-2

# 3. Update staging configuration with secret references
# (Don't hardcode secrets in config files!)

# 4. Test secret retrieval
aws secretsmanager get-secret-value \
  --secret-id temporal/staging/db-password \
  --region us-west-2 \
  --query SecretString \
  --output text
```

**Checklist:**
- [ ] Database passwords generated and stored
- [ ] JWT signing keys generated and stored
- [ ] TLS certificates uploaded to secrets manager
- [ ] Secret retrieval tested
- [ ] Configuration updated with secret references (not plaintext!)

---

### Day 3: Staging Environment Setup

#### Morning (2 hours)
**🏗️ Staging Infrastructure Verification**

```bash
# 1. Verify staging Kubernetes cluster
kubectl cluster-info
kubectl get nodes

# Expected: All nodes Ready

# 2. Verify namespaces
kubectl get namespace temporal-staging || kubectl create namespace temporal-staging

# 3. Verify database accessibility
# PostgreSQL example:
kubectl run -it --rm debug --image=postgres:12 --restart=Never -- \
  psql -h postgresql.staging.example.com -U temporal -d temporal -c "SELECT version();"

# Expected: PostgreSQL 12.x or higher

# 4. Verify network policies (if applicable)
kubectl get networkpolicies -n temporal-staging
```

**Checklist:**
- [ ] Kubernetes cluster healthy
- [ ] Namespace created: `temporal-staging`
- [ ] Database accessible from cluster
- [ ] Database schema deployed (if not already)
- [ ] Network policies configured (if applicable)

#### Afternoon (4 hours)
**📦 Deploy Staging Configuration**

```bash
# 1. Create Kubernetes secrets from secrets manager
kubectl create secret generic temporal-tls-certs \
  --from-file=ca.pem=/tmp/temporal-certs/ca.pem \
  --from-file=frontend-server.pem=/tmp/temporal-certs/frontend-server.pem \
  --from-file=frontend-server-key.pem=/tmp/temporal-certs/frontend-server-key.pem \
  --from-file=client.pem=/tmp/temporal-certs/client.pem \
  --from-file=client-key.pem=/tmp/temporal-certs/client-key.pem \
  -n temporal-staging

# 2. Create ConfigMap from staging config
kubectl create configmap temporal-config \
  --from-file=config.yaml=/home/user/temporal/config/staging-security-enabled.yaml \
  -n temporal-staging

# 3. Verify secrets and config
kubectl get secrets -n temporal-staging
kubectl get configmaps -n temporal-staging

# 4. Test configuration validity (dry-run)
# temporal server config validate --config=/etc/temporal/config/config.yaml
```

**Checklist:**
- [ ] Kubernetes secrets created
- [ ] ConfigMap created
- [ ] Configuration validated (no syntax errors)
- [ ] Secrets accessible by pods

---

### Day 4: Rollback Plan & Testing

#### Morning (3 hours)
**🔙 Rollback Procedure Documentation**

```bash
# Create rollback script
cat > /tmp/rollback-staging.sh << 'EOF'
#!/bin/bash
set -euo pipefail

echo "🔙 Starting rollback to previous configuration..."

# 1. Revert to previous ConfigMap
kubectl rollout undo deployment/temporal-frontend -n temporal-staging
kubectl rollout undo deployment/temporal-history -n temporal-staging
kubectl rollout undo deployment/temporal-matching -n temporal-staging
kubectl rollout undo deployment/temporal-worker -n temporal-staging

# 2. Wait for rollout to complete
kubectl rollout status deployment/temporal-frontend -n temporal-staging --timeout=5m
kubectl rollout status deployment/temporal-history -n temporal-staging --timeout=5m
kubectl rollout status deployment/temporal-matching -n temporal-staging --timeout=5m
kubectl rollout status deployment/temporal-worker -n temporal-staging --timeout=5m

# 3. Verify services healthy
kubectl get pods -n temporal-staging

echo "✅ Rollback complete!"
echo "Verifying service health..."

# Test health endpoints
curl -f http://frontend:7243/health || echo "❌ Frontend health check failed"

echo "Rollback verification complete."
EOF

chmod +x /tmp/rollback-staging.sh
```

**Checklist:**
- [ ] Rollback script created and tested (dry-run)
- [ ] Rollback procedure documented
- [ ] Rollback owner assigned: __________
- [ ] Rollback success criteria defined
- [ ] Team trained on rollback procedure

#### Afternoon (3 hours)
**🧪 Rollback Testing (in dev environment)**

```bash
# Test rollback in dev environment
# 1. Deploy "bad" configuration
# 2. Trigger rollback
# 3. Verify rollback successful
# 4. Measure rollback time (target: <5 minutes)

# Document results
echo "Rollback test results:" > /tmp/rollback-test-results.txt
echo "- Time to rollback: _____ minutes" >> /tmp/rollback-test-results.txt
echo "- Services recovered: ___/4" >> /tmp/rollback-test-results.txt
echo "- Data loss: Yes/No" >> /tmp/rollback-test-results.txt
```

**Checklist:**
- [ ] Rollback tested in dev environment
- [ ] Rollback time measured: _____ minutes (target: <5 min)
- [ ] No data loss during rollback
- [ ] All services recovered successfully
- [ ] Rollback procedure validated

---

### Day 5: Prometheus Alerting Setup

#### Morning (3 hours)
**📊 Deploy Prometheus Alerts**

```bash
# 1. Verify Prometheus installed
kubectl get pods -l app=prometheus -n monitoring

# 2. Deploy alerting rules
kubectl apply -f /home/user/temporal/config/prometheus-alerts.yml

# 3. Verify rules loaded
curl -s http://prometheus:9090/api/v1/rules | jq '.data.groups[].name'

# Expected: temporal_security_critical, temporal_security_high, etc.

# 4. Check for any syntax errors
curl -s http://prometheus:9090/api/v1/rules | jq '.data.groups[].rules[] | select(.health != "ok")'

# Expected: Empty (no unhealthy rules)
```

**Checklist:**
- [ ] Prometheus operational in monitoring namespace
- [ ] 35 alerting rules deployed:
  - [ ] 4 critical security alerts
  - [ ] 7 high security alerts
  - [ ] 3 medium security alerts
  - [ ] 4 critical operational alerts
  - [ ] 6 high operational alerts
  - [ ] 11 medium operational alerts
- [ ] All rules health: OK
- [ ] No syntax errors

#### Afternoon (3 hours)
**🔔 Alertmanager Configuration**

```bash
# 1. Configure Alertmanager routing
cat > /tmp/alertmanager-config.yaml << 'EOF'
route:
  group_by: ['alertname', 'cluster', 'service']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 12h
  receiver: 'team-pager'
  routes:
    - match:
        severity: critical
      receiver: 'pagerduty-critical'
      continue: true
    - match:
        severity: high
      receiver: 'pagerduty-high'
      continue: true
    - match:
        severity: warning
      receiver: 'slack-warnings'

receivers:
  - name: 'pagerduty-critical'
    pagerduty_configs:
      - service_key: '<PAGERDUTY_SERVICE_KEY>'
        severity: 'critical'

  - name: 'pagerduty-high'
    pagerduty_configs:
      - service_key: '<PAGERDUTY_SERVICE_KEY>'
        severity: 'high'

  - name: 'slack-warnings'
    slack_configs:
      - api_url: '<SLACK_WEBHOOK_URL>'
        channel: '#temporal-alerts'
        title: 'Temporal Alert'
        text: '{{ range .Alerts }}{{ .Annotations.description }}{{ end }}'

  - name: 'team-pager'
    pagerduty_configs:
      - service_key: '<PAGERDUTY_SERVICE_KEY>'
EOF

# 2. Apply configuration
kubectl create configmap alertmanager-config \
  --from-file=alertmanager.yml=/tmp/alertmanager-config.yaml \
  -n monitoring \
  --dry-run=client -o yaml | kubectl apply -f -

# 3. Reload Alertmanager
kubectl rollout restart deployment/alertmanager -n monitoring
```

**Checklist:**
- [ ] Alertmanager configured
- [ ] PagerDuty integration tested
- [ ] Slack integration tested
- [ ] Alert routing verified
- [ ] Test alert sent and received

---

### Day 6: Grafana Dashboard Creation

#### All Day (6 hours)
**📈 Security Dashboard Setup**

```bash
# 1. Access Grafana
kubectl port-forward svc/grafana 3000:3000 -n monitoring

# Navigate to http://localhost:3000

# 2. Create dashboard from template
# Download template: https://grafana.com/grafana/dashboards/temporal-security

# 3. Add panels for 12 security metrics:
# - service_authentication_failures
# - service_authorization_failures
# - service_rate_limit_exceeded
# - service_tls_handshake_failures
# - service_jwt_validation_failures
# - service_namespace_isolation_violations
# - service_config_validation_errors
# - service_audit_log_events
# - service_cert_expiry_days
# - service_secret_rotation_age_days
# - service_mTLS_verification_failures
# - service_security_scanner_findings

# 4. Configure alerting in dashboard panels

# 5. Export dashboard JSON
# Save to: /home/user/temporal/config/grafana-security-dashboard.json
```

**Checklist:**
- [ ] Grafana dashboard created
- [ ] All 12 security metrics displayed
- [ ] Alert thresholds configured
- [ ] Dashboard exported (JSON)
- [ ] Dashboard shared with team
- [ ] Team trained on dashboard usage

---

### Day 7: Team Training & Final Prep

#### Morning (3 hours)
**👥 Team Training Session**

**Training Agenda:**
1. Overview of security features (30 min)
   - TLS 1.3, mTLS, JWT auth
   - Rate limiting, audit logging
   - Certificate pinning, timeout enforcement

2. Hands-on: Prometheus alerts (30 min)
   - Viewing active alerts
   - Understanding alert severity
   - Responding to alerts

3. Hands-on: Grafana dashboard (30 min)
   - Navigating security dashboard
   - Interpreting metrics
   - Creating custom queries

4. Rollback procedures (30 min)
   - When to rollback
   - How to execute rollback
   - Verification steps

5. Q&A (30 min)

**Checklist:**
- [ ] Training session completed
- [ ] All team members attended
- [ ] Training materials distributed
- [ ] Questions documented and answered
- [ ] Team confidence level: High/Medium/Low

#### Afternoon (3 hours)
**✅ Final Readiness Check**

```bash
# Run comprehensive pre-deployment check
cat > /tmp/week1-readiness-check.sh << 'EOF'
#!/bin/bash

echo "🔍 Week 1 Readiness Check"
echo "=========================="

# Configuration
echo "✓ Checking configuration files..."
test -f /home/user/temporal/config/staging-security-enabled.yaml && echo "  ✓ Staging config present" || echo "  ✗ Missing staging config"

# Certificates
echo "✓ Checking TLS certificates..."
kubectl get secret temporal-tls-certs -n temporal-staging &>/dev/null && echo "  ✓ TLS secrets created" || echo "  ✗ Missing TLS secrets"

# Monitoring
echo "✓ Checking Prometheus alerts..."
curl -s http://prometheus:9090/api/v1/rules | grep -q "temporal_security" && echo "  ✓ Alerts deployed" || echo "  ✗ Alerts not deployed"

# Dashboard
echo "✓ Checking Grafana dashboard..."
test -f /home/user/temporal/config/grafana-security-dashboard.json && echo "  ✓ Dashboard exported" || echo "  ✗ Dashboard not exported"

# Rollback
echo "✓ Checking rollback procedure..."
test -f /tmp/rollback-staging.sh && test -x /tmp/rollback-staging.sh && echo "  ✓ Rollback script ready" || echo "  ✗ Rollback script missing"

echo ""
echo "=========================="
echo "Week 1 Readiness: COMPLETE"
EOF

chmod +x /tmp/week1-readiness-check.sh
/tmp/week1-readiness-check.sh
```

**Final Checklist:**
- [ ] **Configuration:**
  - [ ] Staging config finalized
  - [ ] All secrets created and accessible
  - [ ] TLS certificates valid and uploaded

- [ ] **Monitoring:**
  - [ ] 35 Prometheus alerts deployed
  - [ ] Alertmanager configured and tested
  - [ ] Grafana dashboard created and shared

- [ ] **Procedures:**
  - [ ] Rollback procedure documented and tested
  - [ ] Team trained on all procedures
  - [ ] Emergency contacts documented

- [ ] **Team Readiness:**
  - [ ] All team members trained
  - [ ] Owners assigned for Week 2
  - [ ] Deployment window reserved
  - [ ] Go/No-Go decision criteria defined

**Go/No-Go Decision for Week 2:**
☐ GO - Proceed with Week 2 staging deployment
☐ NO-GO - Address blockers: _________________

---

## 🎯 Week 1 Success Criteria

✅ **COMPLETE if ALL criteria met:**
- [ ] Team aligned on 30-day roadmap
- [ ] Staging environment fully prepared
- [ ] TLS certificates generated and valid (>90 days)
- [ ] All secrets stored securely (no plaintext)
- [ ] 35 Prometheus alerts deployed and tested
- [ ] Grafana security dashboard operational
- [ ] Rollback procedure tested (< 5 minutes)
- [ ] Team trained on new features
- [ ] Week 2 deployment window scheduled

---

## 📞 Support & Escalation

**Questions during Week 1?**
- Technical Lead: __________
- Security Lead: __________
- SRE Lead: __________

**Blockers?**
- Escalate to: Engineering Manager
- Timeline: Within 24 hours

---

## 🔜 Next: Week 2 Staging Deployment

**Preparation for Week 2:**
1. Review DEPLOYMENT_VERIFICATION_CHECKLIST.md
2. Schedule deployment window (Day 8-9)
3. Brief on-call team
4. Prepare rollback plan
5. Set up war room (Zoom/Slack channel)

**Week 2 Kickoff:**
- Date: __________ (Day 8)
- Time: __________
- Location: [Zoom link / Room]
- Attendees: Security, DevOps, SRE, Platform

---

**Week 1 Sign-off:**
- Completed by: _________________ Date: _______
- Verified by: _________________ Date: _______
- Approved for Week 2: _________________ Date: _______

**Reference:**
- PROJECT_IMPLEMENTATION_ROADMAP.md (Week 1)
- DEPLOYMENT_VERIFICATION_CHECKLIST.md (Pre-Deployment)
- config/staging-security-enabled.yaml
- config/prometheus-alerts.yml
