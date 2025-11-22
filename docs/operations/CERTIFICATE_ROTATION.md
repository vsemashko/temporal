# TLS Certificate Rotation Guide

**Document Version:** 1.0
**Last Updated:** 2025-11-22
**Target Audience:** DevOps Engineers, SREs, Platform Engineers

---

## Table of Contents

1. [Overview](#overview)
2. [Prerequisites](#prerequisites)
3. [Certificate Rotation Strategy](#certificate-rotation-strategy)
4. [Zero-Downtime Rotation Procedure](#zero-downtime-rotation-procedure)
5. [Emergency Rotation (Compromised Certificate)](#emergency-rotation-compromised-certificate)
6. [Automated Rotation](#automated-rotation)
7. [Monitoring and Validation](#monitoring-and-validation)
8. [Troubleshooting](#troubleshooting)
9. [Rollback Procedures](#rollback-procedures)

---

## Overview

This guide covers the procedures for rotating TLS certificates in Temporal Server with zero downtime. Certificate rotation is necessary for:

- **Regular Expiration**: Certificates approaching expiration date
- **Security Compliance**: Meeting organizational security policies
- **Compromise Response**: Responding to potential certificate compromise
- **Algorithm Updates**: Upgrading to stronger cryptographic algorithms

### What Gets Rotated

Temporal Server uses multiple certificate types:

1. **Frontend Server Certificates** - Client-facing gRPC/HTTP endpoints
2. **Internode Server Certificates** - Service-to-service communication
3. **Client Certificates** - For mutual TLS (mTLS) authentication
4. **Database Client Certificates** - For database connections

---

## Prerequisites

Before starting certificate rotation:

### Required Access
- [ ] SSH/kubectl access to all Temporal Server instances
- [ ] Access to certificate management system (internal CA, Let's Encrypt, etc.)
- [ ] Access to configuration management (Git, ConfigMap, etc.)
- [ ] Monitoring system access (metrics, logs)

### Required Information
- [ ] Current certificate expiration dates
- [ ] Certificate Subject Alternative Names (SANs)
- [ ] Current TLS configuration locations
- [ ] Number of Temporal Server instances
- [ ] Deployment method (bare metal, Kubernetes, Docker)

### Pre-Rotation Checklist
- [ ] New certificates generated and validated
- [ ] Certificate chain complete (root + intermediate CAs)
- [ ] Private keys securely stored
- [ ] Backup of current certificates taken
- [ ] Maintenance window scheduled (for emergency rotation)
- [ ] Rollback plan documented
- [ ] Team members notified

---

## Certificate Rotation Strategy

Temporal Server supports **zero-downtime certificate rotation** through certificate refresh mechanisms.

### How It Works

1. **Certificate Refresh Interval**: Temporal periodically reloads certificates from disk
2. **Dual-Certificate Period**: Both old and new certificates are valid during rotation
3. **Gradual Rollout**: Update certificates instance by instance
4. **Client Compatibility**: Clients continue working with both old and new certs

### Configuration for Rotation

Ensure your TLS configuration includes refresh settings:

```yaml
global:
  tls:
    # Certificate refresh interval (REQUIRED for zero-downtime rotation)
    refreshInterval: 1h

    # Expiration monitoring (RECOMMENDED)
    expirationChecks:
      warningWindow: 720h    # 30 days
      errorWindow: 168h      # 7 days
      checkInterval: 12h

    frontend:
      server:
        certFile: /etc/temporal/certs/frontend-server.pem
        keyFile: /etc/temporal/certs/frontend-server-key.pem
        clientCAFiles:
          - /etc/temporal/certs/ca.pem
        requireClientAuth: true
```

**Important**: If `refreshInterval` is not set, you'll need to restart services for certificate changes.

---

## Zero-Downtime Rotation Procedure

### Phase 1: Update CA Certificate (If Needed)

If you're changing the Certificate Authority:

1. **Add New CA to Trust Store**

   Update configuration to trust BOTH old and new CAs:

   ```yaml
   frontend:
     server:
       clientCAFiles:
         - /etc/temporal/certs/old-ca.pem
         - /etc/temporal/certs/new-ca.pem  # Add new CA
   ```

2. **Deploy CA Configuration**

   ```bash
   # Copy new CA to all nodes
   for host in temporal-{1..5}; do
     scp new-ca.pem $host:/etc/temporal/certs/
   done

   # Update config (using your deployment tool)
   kubectl apply -f temporal-config.yaml
   # OR
   ansible-playbook update-temporal-config.yml
   ```

3. **Verify CA Trust**

   ```bash
   # Check logs for CA reload
   grep "reloaded TLS" /var/log/temporal/temporal.log

   # Verify multiple CAs are trusted
   openssl s_client -connect temporal.example.com:7233 \
     -CAfile /etc/temporal/certs/old-ca.pem
   openssl s_client -connect temporal.example.com:7233 \
     -CAfile /etc/temporal/certs/new-ca.pem
   ```

### Phase 2: Generate New Certificates

1. **Create Certificate Signing Request (CSR)**

   ```bash
   # Generate private key
   openssl genrsa -out frontend-server-new-key.pem 2048

   # Create CSR with SANs
   cat > csr.conf <<EOF
   [req]
   default_bits = 2048
   prompt = no
   default_md = sha256
   distinguished_name = dn
   req_extensions = v3_req

   [dn]
   C=US
   ST=California
   L=San Francisco
   O=Example Corp
   CN=temporal.example.com

   [v3_req]
   keyUsage = keyEncipherment, dataEncipherment
   extendedKeyUsage = serverAuth
   subjectAltName = @alt_names

   [alt_names]
   DNS.1 = temporal.example.com
   DNS.2 = *.temporal.example.com
   DNS.3 = temporal-frontend-1.internal
   DNS.4 = temporal-frontend-2.internal
   IP.1 = 10.0.1.100
   EOF

   openssl req -new -key frontend-server-new-key.pem \
     -out frontend-server-new.csr -config csr.conf
   ```

2. **Sign Certificate**

   ```bash
   # Sign with your CA
   openssl x509 -req -in frontend-server-new.csr \
     -CA ca.pem -CAkey ca-key.pem -CAcreateserial \
     -out frontend-server-new.pem -days 365 \
     -extensions v3_req -extfile csr.conf
   ```

3. **Validate New Certificate**

   ```bash
   # Check expiration date
   openssl x509 -in frontend-server-new.pem -noout -dates

   # Verify certificate chain
   openssl verify -CAfile ca.pem frontend-server-new.pem

   # Check SANs
   openssl x509 -in frontend-server-new.pem -noout -text | grep -A1 "Subject Alternative Name"

   # Verify key matches certificate
   diff <(openssl x509 -in frontend-server-new.pem -noout -modulus) \
        <(openssl rsa -in frontend-server-new-key.pem -noout -modulus)
   ```

### Phase 3: Deploy New Certificates (Rolling Update)

**For Kubernetes Deployments:**

```bash
# Update secret with new certificates
kubectl create secret generic temporal-certs \
  --from-file=frontend-server.pem=frontend-server-new.pem \
  --from-file=frontend-server-key.pem=frontend-server-new-key.pem \
  --from-file=ca.pem=ca.pem \
  --dry-run=client -o yaml | kubectl apply -f -

# Trigger cert reload (if using refreshInterval)
# Certificates will be reloaded automatically within refreshInterval

# OR force pod restart for immediate reload
kubectl rollout restart deployment/temporal-frontend
kubectl rollout status deployment/temporal-frontend
```

**For Bare Metal/VM Deployments:**

```bash
#!/bin/bash
# rotate-certs.sh - Rolling certificate update script

HOSTS="temporal-1 temporal-2 temporal-3 temporal-4 temporal-5"
CERT_DIR="/etc/temporal/certs"
BACKUP_DIR="/etc/temporal/certs/backup-$(date +%Y%m%d-%H%M%S)"

for host in $HOSTS; do
  echo "=== Rotating certificates on $host ==="

  # Create backup
  ssh $host "mkdir -p $BACKUP_DIR && cp $CERT_DIR/*.pem $BACKUP_DIR/"

  # Copy new certificates
  scp frontend-server-new.pem $host:$CERT_DIR/frontend-server.pem
  scp frontend-server-new-key.pem $host:$CERT_DIR/frontend-server-key.pem

  # Set permissions
  ssh $host "chmod 644 $CERT_DIR/frontend-server.pem && \
             chmod 400 $CERT_DIR/frontend-server-key.pem && \
             chown temporal:temporal $CERT_DIR/frontend-server*.pem"

  # Wait for auto-reload (based on refreshInterval)
  echo "Waiting 65 seconds for certificate refresh..."
  sleep 65

  # Verify reload
  ssh $host "grep 'reloaded TLS' /var/log/temporal/temporal.log | tail -1"

  # Test connectivity
  echo "Testing TLS connectivity..."
  openssl s_client -connect $host:7233 -CAfile ca.pem < /dev/null 2>&1 | \
    grep -E "(Verify return code|subject=)"

  if [ $? -ne 0 ]; then
    echo "ERROR: Certificate rotation failed on $host"
    echo "Rolling back..."
    ssh $host "cp $BACKUP_DIR/*.pem $CERT_DIR/"
    exit 1
  fi

  echo "SUCCESS: $host certificate rotated"
  echo ""

  # Wait before next host (stagger rollout)
  sleep 30
done

echo "=== Certificate rotation complete on all hosts ==="
```

### Phase 4: Verify Rotation

1. **Check Certificate Expiration**

   ```bash
   # Verify new certificate is in use
   echo | openssl s_client -connect temporal.example.com:7233 2>/dev/null | \
     openssl x509 -noout -dates
   ```

2. **Test Client Connectivity**

   ```bash
   # Test with Temporal CLI
   temporal operator cluster health \
     --address temporal.example.com:7233 \
     --tls-cert-path client.pem \
     --tls-key-path client-key.pem \
     --tls-ca-path ca.pem

   # Test with curl (for HTTP API)
   curl -v --cacert ca.pem https://temporal.example.com:7243/health
   ```

3. **Monitor Metrics**

   ```promql
   # Check for TLS errors
   rate(temporal_tls_errors_total[5m])

   # Monitor connection success rate
   rate(temporal_grpc_requests_total{code="OK"}[5m]) /
   rate(temporal_grpc_requests_total[5m])
   ```

4. **Check Logs**

   ```bash
   # Look for successful TLS reloads
   grep "reloaded TLS" /var/log/temporal/temporal.log

   # Check for TLS errors
   grep -i "tls\|certificate" /var/log/temporal/temporal.log | grep -i error
   ```

### Phase 5: Cleanup Old CA (If Applicable)

After confirming all services use the new certificate:

1. **Wait for Old Certificate Expiry + Buffer**

   Wait at least 7 days after old certificate expiration to ensure no cached certificates remain.

2. **Remove Old CA from Trust Store**

   ```yaml
   frontend:
     server:
       clientCAFiles:
         - /etc/temporal/certs/new-ca.pem  # Only new CA now
   ```

3. **Remove Old Certificate Files**

   ```bash
   # Archive old certificates (don't delete immediately)
   mkdir -p /etc/temporal/certs/archive-$(date +%Y%m%d)
   mv /etc/temporal/certs/backup-* /etc/temporal/certs/archive-$(date +%Y%m%d)/
   ```

---

## Emergency Rotation (Compromised Certificate)

If a certificate is compromised, follow this expedited procedure:

### Immediate Actions (Within 1 Hour)

1. **Revoke Compromised Certificate**

   ```bash
   # Add to CRL (Certificate Revocation List)
   openssl ca -revoke compromised-cert.pem -keyfile ca-key.pem -cert ca.pem

   # Update CRL
   openssl ca -gencrl -keyfile ca-key.pem -cert ca.pem -out crl.pem

   # Distribute CRL to all instances
   for host in temporal-{1..5}; do
     scp crl.pem $host:/etc/temporal/certs/
   done
   ```

2. **Generate Emergency Certificates**

   Use the same process as Phase 2, but with immediate signing.

3. **Force Immediate Rotation**

   ```bash
   # Disable automatic refresh, restart services immediately
   for host in temporal-{1..5}; do
     # Copy new certs
     scp emergency-*.pem $host:/etc/temporal/certs/

     # Restart service
     ssh $host "systemctl restart temporal-frontend"
   done
   ```

4. **Notify Stakeholders**

   - Security team
   - Clients/partners using mTLS
   - Compliance team

### Post-Incident (Within 24 Hours)

1. **Investigate Compromise**
   - Review access logs
   - Identify how compromise occurred
   - Implement additional controls

2. **Update Documentation**
   - Document incident timeline
   - Update runbooks with lessons learned

---

## Automated Rotation

### Using cert-manager (Kubernetes)

```yaml
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: temporal-frontend-cert
  namespace: temporal
spec:
  secretName: temporal-frontend-tls
  duration: 2160h  # 90 days
  renewBefore: 360h  # 15 days before expiration
  issuerRef:
    name: temporal-ca-issuer
    kind: ClusterIssuer
  dnsNames:
    - temporal.example.com
    - "*.temporal.example.com"
```

### Using Custom Scripts

```bash
#!/bin/bash
# auto-rotate.sh - Automated certificate rotation

# Configuration
DAYS_BEFORE_EXPIRY=30
CERT_FILE="/etc/temporal/certs/frontend-server.pem"
CA_SERVER="ca.example.com"

# Check expiration
expiry_date=$(openssl x509 -in $CERT_FILE -noout -enddate | cut -d= -f2)
expiry_epoch=$(date -d "$expiry_date" +%s)
now_epoch=$(date +%s)
days_remaining=$(( ($expiry_epoch - $now_epoch) / 86400 ))

if [ $days_remaining -lt $DAYS_BEFORE_EXPIRY ]; then
  echo "Certificate expires in $days_remaining days, rotating..."

  # Generate new cert (implement your CA integration)
  generate_new_certificate

  # Deploy via your automation
  deploy_certificate

  # Send notification
  send_notification "Certificate rotated successfully. $days_remaining days remained."
else
  echo "Certificate valid for $days_remaining days, no rotation needed"
fi
```

Add to crontab:
```cron
# Check certificate expiration daily at 2 AM
0 2 * * * /usr/local/bin/auto-rotate.sh >> /var/log/temporal/cert-rotation.log 2>&1
```

---

## Monitoring and Validation

### Expiration Monitoring

```promql
# Alert when certificate expires in < 30 days
(temporal_tls_certificate_expiry_timestamp - time()) / 86400 < 30
```

**Alert Configuration:**

```yaml
- alert: TLSCertificateExpiringSoon
  expr: (temporal_tls_certificate_expiry_timestamp - time()) / 86400 < 30
  for: 1h
  labels:
    severity: warning
  annotations:
    summary: "TLS certificate expiring soon"
    description: "Certificate for {{ $labels.instance }} expires in {{ $value }} days"

- alert: TLSCertificateExpiring
  expr: (temporal_tls_certificate_expiry_timestamp - time()) / 86400 < 7
  for: 1h
  labels:
    severity: critical
  annotations:
    summary: "TLS certificate expiring in < 7 days"
```

### Rotation Validation Script

```bash
#!/bin/bash
# validate-rotation.sh

ENDPOINT="temporal.example.com:7233"
CA_FILE="/etc/temporal/certs/ca.pem"

echo "=== TLS Certificate Validation ==="

# Get certificate
cert_info=$(echo | openssl s_client -connect $ENDPOINT -CAfile $CA_FILE 2>/dev/null)

# Extract expiration
expiry=$(echo "$cert_info" | openssl x509 -noout -enddate | cut -d= -f2)
echo "Expires: $expiry"

# Calculate days remaining
expiry_epoch=$(date -d "$expiry" +%s)
days_remaining=$(( ($expiry_epoch - $(date +%s)) / 86400 ))
echo "Days remaining: $days_remaining"

# Verify chain
echo "$cert_info" | openssl x509 -noout -text | grep -A1 "Issuer:"

# Check SANs
echo "Subject Alternative Names:"
echo "$cert_info" | openssl x509 -noout -text | grep -A1 "Subject Alternative Name"

# Verify return code
return_code=$(echo "$cert_info" | grep "Verify return code" | cut -d: -f2 | tr -d ' ')
if [ "$return_code" == "0 (ok)" ]; then
  echo "✓ Certificate validation: PASSED"
  exit 0
else
  echo "✗ Certificate validation: FAILED ($return_code)"
  exit 1
fi
```

---

## Troubleshooting

### Issue: "Certificate has expired"

**Symptoms:**
```
TLS handshake failed: certificate has expired
```

**Resolution:**
1. Check if rotation completed on all instances
   ```bash
   for host in temporal-{1..5}; do
     echo "=== $host ==="
     ssh $host "openssl x509 -in /etc/temporal/certs/frontend-server.pem -noout -dates"
   done
   ```

2. Force immediate rotation on instances with expired certs
3. Check `refreshInterval` is configured

### Issue: "Certificate chain incomplete"

**Symptoms:**
```
TLS handshake failed: unable to get local issuer certificate
```

**Resolution:**
1. Verify CA file includes full chain (root + intermediate)
   ```bash
   # Should show multiple certificates
   grep -c "BEGIN CERTIFICATE" /etc/temporal/certs/ca.pem
   ```

2. Rebuild CA bundle:
   ```bash
   cat intermediate-ca.pem root-ca.pem > ca.pem
   ```

### Issue: "Certificate name mismatch"

**Symptoms:**
```
TLS handshake failed: x509: certificate is valid for temporal-old.example.com, not temporal.example.com
```

**Resolution:**
1. Verify SANs in new certificate match current hostname
2. Regenerate certificate with correct SANs

### Issue: Rotation caused service disruption

**Resolution:**
1. Check if `refreshInterval` was configured
2. If not, rotation requires restart
3. Implement rolling restart procedure for next rotation

---

## Rollback Procedures

If certificate rotation causes issues:

### Quick Rollback (< 5 minutes)

```bash
#!/bin/bash
# rollback-certs.sh

HOSTS="temporal-1 temporal-2 temporal-3 temporal-4 temporal-5"
BACKUP_DIR="/etc/temporal/certs/backup-YYYYMMDD-HHMMSS"  # Use actual timestamp

for host in $HOSTS; do
  echo "Rolling back certificates on $host"
  ssh $host "cp $BACKUP_DIR/*.pem /etc/temporal/certs/"

  # Wait for auto-reload or force restart
  ssh $host "systemctl restart temporal-frontend"
done
```

### Verification After Rollback

```bash
# Verify old certificates restored
for host in temporal-{1..5}; do
  echo "=== Checking $host ==="
  ssh $host "openssl x509 -in /etc/temporal/certs/frontend-server.pem -noout -subject"
done

# Test connectivity
temporal operator cluster health
```

---

## Best Practices

1. **Always Configure refreshInterval**
   - Enables zero-downtime rotation
   - Set to 1 hour for production

2. **Maintain Certificate Inventory**
   - Track all certificates and expiration dates
   - Use monitoring to prevent surprise expirations

3. **Test in Staging First**
   - Always test rotation procedure in non-production environment
   - Validate scripts and automation

4. **Keep Backups**
   - Always backup before rotation
   - Keep backups for 90 days minimum

5. **Stagger Rotation**
   - Rotate one instance at a time
   - Wait for verification before proceeding

6. **Document Everything**
   - Keep rotation logs
   - Document any issues and resolutions

7. **Automate Where Possible**
   - Use cert-manager or similar tools
   - Automate expiration monitoring

8. **Regular Drills**
   - Practice rotation quarterly
   - Practice emergency rotation annually

---

## Related Documentation

- [SECURITY_OPERATOR_GUIDE.md](../../SECURITY_OPERATOR_GUIDE.md) - TLS configuration
- [JWT_KEY_ROTATION.md](./JWT_KEY_ROTATION.md) - JWT key rotation procedures
- [SECRETS_ROTATION_RUNBOOK.md](./SECRETS_ROTATION_RUNBOOK.md) - Quick reference guide

---

**Document Owner:** Security & Infrastructure Team
**Last Reviewed:** 2025-11-22
**Next Review:** 2026-02-22 (quarterly)
