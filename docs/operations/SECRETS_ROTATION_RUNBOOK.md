# Secrets Rotation Runbook - Quick Reference

**Document Version:** 1.0
**Last Updated:** 2025-11-22
**Format:** Quick Reference Guide

---

## 🚨 Emergency Quick Links

| Scenario | Go To Section |
|----------|---------------|
| 🔥 Certificate compromised | [Emergency Certificate Rotation](#emergency-certificate-rotation) |
| 🔥 JWT key compromised | [Emergency JWT Key Rotation](#emergency-jwt-key-rotation) |
| ✅ Planned certificate rotation | [Standard Certificate Rotation](#standard-certificate-rotation) |
| ✅ Planned JWT key rotation | [Standard JWT Key Rotation](#standard-jwt-key-rotation) |
| ❓ Rotation validation | [Validation Checks](#validation-checks) |
| 🔙 Need to rollback | [Rollback Procedures](#rollback-procedures) |

---

## Standard Certificate Rotation

### Timeline: 5-7 days

```
Day 0: Add new CA (if changing CA)
Day 1: Generate and deploy new certificates
Day 2-6: Validate and monitor
Day 7: Remove old CA
```

### Quick Commands

```bash
# 1. Generate new certificate
openssl genrsa -out server-new-key.pem 2048
openssl req -new -key server-new-key.pem -out server-new.csr -config csr.conf
openssl x509 -req -in server-new.csr -CA ca.pem -CAkey ca-key.pem \
  -CAcreateserial -out server-new.pem -days 365 -extensions v3_req -extfile csr.conf

# 2. Validate certificate
openssl x509 -in server-new.pem -noout -dates
openssl verify -CAfile ca.pem server-new.pem

# 3. Deploy to Kubernetes
kubectl create secret generic temporal-certs \
  --from-file=server.pem=server-new.pem \
  --from-file=server-key.pem=server-new-key.pem \
  --dry-run=client -o yaml | kubectl apply -f -

# 4. Wait for auto-reload (if refreshInterval configured)
sleep 65

# 5. Verify
openssl s_client -connect temporal.example.com:7233 -CAfile ca.pem < /dev/null

# 6. After 7 days, remove old CA from configuration
```

### Pre-Flight Checklist

- [ ] `refreshInterval` configured in TLS config
- [ ] New certificate generated and validated
- [ ] Backup of current certificates taken
- [ ] SANs include all required hostnames
- [ ] Certificate chain complete

### Verification Steps

```bash
# Check certificate is loaded
echo | openssl s_client -connect temporal.example.com:7233 2>/dev/null | \
  openssl x509 -noout -subject -dates

# Test client connectivity
temporal operator cluster health --address temporal.example.com:7233

# Check logs
grep "reloaded TLS" /var/log/temporal/temporal.log
```

---

## Standard JWT Key Rotation

### Timeline: 25-48 hours (depends on max token TTL)

```
Hour 0: Add new public key to Temporal
Hour 1: Update IdP to sign with new key
Hour 2-24: Dual-key period (both keys valid)
Hour 25: Remove old public key
```

### Quick Commands

```bash
# 1. Generate new key pair
openssl genrsa -out jwt-private-new.pem 4096
openssl rsa -in jwt-private-new.pem -pubout -out jwt-public-new.pem
KID="temporal-jwt-$(date +%Y%m%d)"

# 2. Add new public key to Temporal (keep old)
kubectl create configmap temporal-jwt-keys \
  --from-file=jwt-public-old.pem \
  --from-file=jwt-public-new.pem \
  --dry-run=client -o yaml | kubectl apply -f -

# 3. Update IdP to use new private key (your IdP-specific command)
update_idp_signing_key jwt-private-new.pem "$KID"

# 4. Wait for max token TTL + buffer
# If tokens expire in 1 hour: wait 25 hours
sleep $((25 * 3600))

# 5. Remove old key
kubectl create configmap temporal-jwt-keys \
  --from-file=jwt-public-new.pem \
  --dry-run=client -o yaml | kubectl apply -f -

# 6. Verify
TOKEN=$(get_new_token)
temporal workflow list --headers "authorization=Bearer $TOKEN"
```

### Pre-Flight Checklist

- [ ] Maximum token TTL known
- [ ] New key pair generated
- [ ] Key ID (kid) assigned
- [ ] Dual-key period calculated (max TTL + 24h buffer)
- [ ] IdP deployment plan ready

### Verification Steps

```bash
# Check which keys are being used
kubectl logs deployment/temporal-frontend | grep "JWT.*kid=" | \
  grep -oP 'kid=\K[^,]+' | sort | uniq -c

# Decode token to verify kid
echo "$TOKEN" | cut -d. -f1 | base64 -d | jq .kid

# Test token validation
temporal workflow list --headers "authorization=Bearer $TOKEN"
```

---

## Emergency Certificate Rotation

### ⏱️ Target: Complete in < 2 hours

```bash
#!/bin/bash
# emergency-cert-rotation.sh

set -e

echo "🚨 EMERGENCY CERTIFICATE ROTATION"

# 1. Generate emergency certificate (5 min)
openssl genrsa -out emergency-key.pem 2048
openssl req -new -key emergency-key.pem -out emergency.csr -subj "/CN=temporal.example.com"
openssl x509 -req -in emergency.csr -CA ca.pem -CAkey ca-key.pem \
  -CAcreateserial -out emergency.pem -days 365

# 2. Deploy immediately (10 min)
for host in temporal-{1..5}; do
  scp emergency*.pem $host:/etc/temporal/certs/
  ssh $host "systemctl restart temporal-frontend"
done

# 3. Revoke compromised certificate (5 min)
openssl ca -revoke compromised.pem -keyfile ca-key.pem -cert ca.pem
openssl ca -gencrl -out crl.pem -keyfile ca-key.pem -cert ca.pem

# 4. Notify stakeholders
notify_security_team "Emergency certificate rotation complete"

echo "✅ Emergency rotation complete"
```

### Immediate Actions

1. **Revoke compromised certificate** (if CA supports it)
2. **Generate emergency certificate** (can skip SANs validation in emergency)
3. **Force immediate deployment** (restart services, skip auto-reload)
4. **Update CRL** (Certificate Revocation List)
5. **Notify all clients** (especially for mTLS)

### Post-Incident

- [ ] Root cause analysis completed
- [ ] Access logs reviewed
- [ ] Additional controls implemented
- [ ] Incident documentation updated

---

## Emergency JWT Key Rotation

### ⏱️ Target: Complete in < 1 hour

```bash
#!/bin/bash
# emergency-jwt-rotation.sh

set -e

echo "🚨 EMERGENCY JWT KEY ROTATION"

# 1. Generate emergency key (2 min)
openssl genrsa -out jwt-emergency.pem 4096
openssl rsa -in jwt-emergency.pem -pubout -out jwt-emergency-pub.pem
EMERGENCY_KID="temporal-jwt-emergency-$(date +%Y%m%d%H%M)"

# 2. Replace public key in Temporal immediately (5 min)
kubectl create configmap temporal-jwt-keys \
  --from-file=jwt-public.pem=jwt-emergency-pub.pem \
  --dry-run=client -o yaml | kubectl apply -f -

# Force reload
kubectl rollout restart deployment/temporal-frontend

# 3. Deploy to IdP immediately (10 min)
deploy_emergency_key_to_idp jwt-emergency.pem "$EMERGENCY_KID"

# 4. Invalidate all existing tokens (if possible)
invalidate_all_sessions

# 5. Notify users to re-authenticate
notify_users "Please re-login: security incident response"

echo "✅ Emergency JWT key rotation complete"
```

### Immediate Actions

1. **Generate emergency key pair**
2. **Replace public key immediately** (no dual-key period)
3. **Force restart Temporal services**
4. **Deploy to IdP immediately**
5. **Invalidate existing sessions/tokens**
6. **Force user re-authentication**

### Post-Incident

- [ ] Audit key access logs
- [ ] Identify compromise source
- [ ] Implement HSM/KMS for key storage
- [ ] Update security procedures

---

## Validation Checks

### Certificate Validation

```bash
# Quick validation script
validate_certificate() {
  ENDPOINT=$1
  CA_FILE=$2

  echo "=== Validating Certificate on $ENDPOINT ==="

  # Get certificate
  CERT=$(echo | openssl s_client -connect $ENDPOINT -CAfile $CA_FILE 2>/dev/null)

  # Check expiration
  echo "$CERT" | openssl x509 -noout -dates

  # Verify chain
  echo "$CERT" | openssl x509 -noout -text | grep Issuer

  # Check return code
  echo "$CERT" | grep "Verify return code: 0 (ok)"
}

# Usage
validate_certificate "temporal.example.com:7233" "/etc/temporal/certs/ca.pem"
```

### JWT Key Validation

```bash
# Quick JWT validation
validate_jwt() {
  TOKEN=$1
  ENDPOINT=$2

  echo "=== Validating JWT Token ==="

  # Decode header
  HEADER=$(echo "$TOKEN" | cut -d. -f1 | base64 -d 2>/dev/null)
  echo "Token Header: $HEADER"

  # Extract kid
  KID=$(echo "$HEADER" | jq -r '.kid')
  echo "Key ID: $KID"

  # Test with Temporal
  temporal workflow list \
    --address "$ENDPOINT" \
    --headers "authorization=Bearer $TOKEN"

  echo "✅ Token validation successful"
}

# Usage
TOKEN=$(get_new_token)
validate_jwt "$TOKEN" "temporal.example.com:7233"
```

### Health Check Script

```bash
#!/bin/bash
# health-check.sh - Run after rotation

echo "=== Post-Rotation Health Check ==="

# 1. Certificate expiration
echo "Certificate expiration:"
echo | openssl s_client -connect temporal.example.com:7233 2>/dev/null | \
  openssl x509 -noout -dates

# 2. TLS connectivity
echo "TLS connectivity:"
temporal operator cluster health

# 3. JWT authentication
echo "JWT authentication:"
TOKEN=$(get_new_token)
temporal workflow list --headers "authorization=Bearer $TOKEN" | head -5

# 4. Check logs for errors
echo "Recent errors:"
kubectl logs deployment/temporal-frontend --since=1h | grep -i error | tail -10

# 5. Metrics check
echo "Error rate:"
curl -s "http://prometheus:9090/api/v1/query?query=rate(temporal_grpc_errors_total[5m])"

echo "✅ Health check complete"
```

---

## Rollback Procedures

### Certificate Rollback

```bash
# Quick rollback to previous certificate
rollback_certificate() {
  BACKUP_DIR=$1  # e.g., /etc/temporal/certs/backup-20251120

  echo "🔙 Rolling back to backup: $BACKUP_DIR"

  # Restore from backup
  for host in temporal-{1..5}; do
    ssh $host "cp $BACKUP_DIR/*.pem /etc/temporal/certs/"
    ssh $host "systemctl restart temporal-frontend"
  done

  echo "✅ Rollback complete"
}

# Usage
rollback_certificate "/etc/temporal/certs/backup-20251120-1430"
```

### JWT Key Rollback

```bash
# Quick rollback to previous JWT key
rollback_jwt_key() {
  OLD_PUBLIC_KEY=$1

  echo "🔙 Rolling back to old JWT key"

  # Restore old public key
  kubectl create configmap temporal-jwt-keys \
    --from-file=jwt-public.pem="$OLD_PUBLIC_KEY" \
    --dry-run=client -o yaml | kubectl apply -f -

  # Restart to force reload
  kubectl rollout restart deployment/temporal-frontend

  # Revert IdP
  revert_idp_to_old_key

  echo "✅ Rollback complete"
}

# Usage
rollback_jwt_key "/etc/temporal/jwt-keys/archive/jwt-public-old.pem"
```

---

## Common Troubleshooting

### "Certificate has expired"

```bash
# Check all instances
for host in temporal-{1..5}; do
  echo "=== $host ==="
  ssh $host "openssl x509 -in /etc/temporal/certs/server.pem -noout -dates"
done

# Fix: Deploy new certificate to instances still using old cert
```

### "Unknown kid in JWT"

```bash
# Check kid in token
echo "$TOKEN" | cut -d. -f1 | base64 -d | jq .kid

# Check configured kids
kubectl get configmap temporal-jwt-keys -o yaml

# Fix: Ensure kid in token matches configured keys
```

### "TLS handshake failed"

```bash
# Test TLS connection
openssl s_client -connect temporal.example.com:7233 -CAfile ca.pem

# Check for:
# - Certificate chain issues
# - Wrong CA file
# - Certificate/key mismatch

# Fix: Verify certificate chain is complete
cat intermediate-ca.pem root-ca.pem > ca.pem
```

###"Permission denied: authentication failed"

```bash
# Check JWT validation
temporal workflow list --headers "authorization=Bearer $TOKEN" --verbose

# Common issues:
# - Token expired
# - Wrong signing key
# - Public key not loaded

# Fix: Verify public key is loaded and refresh interval passed
kubectl logs deployment/temporal-frontend | grep "JWT"
```

---

## Monitoring Queries

### Prometheus Queries

```promql
# Certificate expiration (days)
(temporal_tls_certificate_expiry_timestamp - time()) / 86400

# JWT validation success rate
rate(temporal_auth_token_validated_total{result="success"}[5m]) /
rate(temporal_auth_token_validated_total[5m])

# Authentication failures
rate(temporal_auth_failure_total[5m])

# TLS errors
rate(temporal_tls_errors_total[5m])
```

### Log Queries

```bash
# Certificate reload events
kubectl logs deployment/temporal-frontend | grep "reloaded TLS"

# JWT validation failures
kubectl logs deployment/temporal-frontend | grep "JWT.*failed"

# Authentication errors
kubectl logs deployment/temporal-frontend | grep "authentication failed"
```

---

## Rotation Schedule Template

| Item | Last Rotated | Next Rotation | Owner | Status |
|------|--------------|---------------|-------|--------|
| Frontend TLS Cert | 2025-11-22 | 2026-11-22 | Platform Team | ✅ Current |
| Internode TLS Cert | 2025-11-22 | 2026-11-22 | Platform Team | ✅ Current |
| JWT Signing Key | 2025-10-15 | 2026-01-15 | Security Team | ⚠️ Due Soon |
| Database Client Cert | 2025-09-01 | 2026-09-01 | Database Team | ✅ Current |
| CA Root Certificate | 2024-01-01 | 2034-01-01 | Security Team | ✅ Current |

---

## Contact Information

### Escalation Path

| Severity | Contact | Response Time |
|----------|---------|---------------|
| 🔴 P0 - Service Down | On-Call Engineer: +1-555-0100 | 15 minutes |
| 🟠 P1 - Certificate Compromised | Security Team: security@example.com | 1 hour |
| 🟡 P2 - Planned Rotation | Platform Team: platform@example.com | 1 business day |
| 🟢 P3 - Questions | Documentation: docs@example.com | 2 business days |

### Key Personnel

- **Security Lead:** security-lead@example.com
- **Platform Lead:** platform-lead@example.com
- **On-Call Rotation:** https://oncall.example.com/temporal

---

## Reference Documentation

| Document | Purpose | Link |
|----------|---------|------|
| Certificate Rotation Guide | Detailed TLS rotation procedures | [CERTIFICATE_ROTATION.md](./CERTIFICATE_ROTATION.md) |
| JWT Key Rotation Guide | Detailed JWT key rotation procedures | [JWT_KEY_ROTATION.md](./JWT_KEY_ROTATION.md) |
| Security Operator Guide | Security configuration reference | [SECURITY_OPERATOR_GUIDE.md](../../SECURITY_OPERATOR_GUIDE.md) |
| Security Roadmap | Future security improvements | [SECURITY_ROADMAP.md](../../SECURITY_ROADMAP.md) |

---

## Automation Scripts

Located in: `/usr/local/bin/temporal-security/`

```
├── auto-rotate-certs.sh          # Automated certificate rotation
├── auto-rotate-jwt-keys.sh       # Automated JWT key rotation
├── emergency-cert-rotation.sh    # Emergency certificate rotation
├── emergency-jwt-rotation.sh     # Emergency JWT key rotation
├── validate-certificate.sh       # Certificate validation
├── validate-jwt-keys.sh          # JWT key validation
├── health-check.sh               # Post-rotation health check
└── rollback-rotation.sh          # Generic rollback script
```

---

**Document Owner:** Security & Infrastructure Team
**Last Updated:** 2025-11-22
**Emergency Hotline:** +1-555-SECURITY

**Print this page and keep near your desk for quick reference!**
