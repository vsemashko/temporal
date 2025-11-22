# JWT Signing Key Rotation Guide

**Document Version:** 1.0
**Last Updated:** 2025-11-22
**Target Audience:** Security Engineers, DevOps Engineers, SREs

---

## Table of Contents

1. [Overview](#overview)
2. [Prerequisites](#prerequisites)
3. [Rotation Strategy](#rotation-strategy)
4. [Zero-Downtime Rotation Procedure](#zero-downtime-rotation-procedure)
5. [Emergency Rotation (Key Compromise)](#emergency-rotation-key-compromise)
6. [Automated Rotation](#automated-rotation)
7. [Monitoring and Validation](#monitoring-and-validation)
8. [Troubleshooting](#troubleshooting)
9. [Rollback Procedures](#rollback-procedures)

---

## Overview

This guide covers procedures for rotating JWT signing keys used for Temporal Server authorization. JWT key rotation is necessary for:

- **Security Compliance**: Regular key rotation per security policy
- **Key Compromise**: Responding to potential key exposure
- **Algorithm Upgrades**: Migrating to stronger signing algorithms
- **Operational Best Practices**: Limiting key lifetime reduces impact of compromise

### JWT Authentication in Temporal

Temporal uses JWT tokens for client authentication with the following flow:

1. Client obtains JWT from identity provider (IdP)
2. Client includes JWT in `authorization` header
3. Temporal validates JWT signature using public key
4. Temporal extracts claims and enforces authorization

### What Gets Rotated

- **JWT Signing Keys**: Private keys used by IdP to sign tokens
- **JWT Public Keys**: Public keys used by Temporal to verify tokens
- **Key IDs (kid)**: Identifier in JWT header to specify which key to use

---

## Prerequisites

Before starting JWT key rotation:

### Required Access
- [ ] Access to JWT signing infrastructure (IdP, auth service)
- [ ] Access to Temporal configuration (ConfigMaps, secrets)
- [ ] Access to monitoring systems

### Required Knowledge
- [ ] Current JWT key configuration
- [ ] Token expiration times (`exp` claim)
- [ ] Maximum token lifetime in your environment
- [ ] Number of JWT signers (may be multiple for HA)

### Pre-Rotation Checklist
- [ ] New key pair generated
- [ ] Key ID (kid) assigned to new key
- [ ] Current token expiration policy understood
- [ ] Maximum overlap period calculated
- [ ] Rollback plan documented
- [ ] Monitoring alerts configured

---

## Rotation Strategy

JWT key rotation uses a **dual-key period** where both old and new keys are valid simultaneously.

### How It Works

```
Timeline:
|-------- Old Key Only --------|-- Dual Key Period --|---- New Key Only -----|
                                ^                      ^
                          Add New Key            Remove Old Key
                          (Day 0)               (Day 0 + Max Token TTL)
```

1. **Add New Key**: Temporal configured to trust BOTH old and new keys
2. **Dual-Key Period**: Tokens signed with either key are valid
3. **IdP Switches**: IdP starts signing new tokens with new key
4. **Wait Period**: Wait for all old tokens to expire
5. **Remove Old Key**: Temporal stops trusting old key

### Key Concepts

**Maximum Token Lifetime**: Longest time a token could be valid
- Set by `exp` claim in JWT
- Example: If tokens expire in 1 hour, max lifetime = 1 hour

**Dual-Key Period Duration**: Must be ≥ maximum token lifetime
- Recommended: Max token lifetime + 24 hours buffer
- Example: If max token lifetime = 1 hour, dual-key period = 25 hours

---

## Zero-Downtime Rotation Procedure

### Phase 1: Generate New Key Pair

1. **Generate RSA Key Pair** (Most Common)

   ```bash
   # Generate 4096-bit RSA private key
   openssl genrsa -out jwt-private-new.pem 4096

   # Extract public key
   openssl rsa -in jwt-private-new.pem -pubout -out jwt-public-new.pem

   # Assign Key ID (kid) - use timestamp or UUID
   KID="temporal-jwt-$(date +%Y%m%d)"
   echo "$KID" > jwt-kid-new.txt
   ```

2. **Or Generate ECDSA Key Pair** (Modern Alternative)

   ```bash
   # Generate EC private key (P-256 curve)
   openssl ecparam -name prime256v1 -genkey -noout -out jwt-private-new.pem

   # Extract public key
   openssl ec -in jwt-private-new.pem -pubout -out jwt-public-new.pem

   # Assign Key ID
   KID="temporal-jwt-ec-$(date +%Y%m%d)"
   echo "$KID" > jwt-kid-new.txt
   ```

3. **Validate Key Pair**

   ```bash
   # Test signing and verification
   echo "test payload" | \
     openssl dgst -sha256 -sign jwt-private-new.pem | \
     openssl dgst -sha256 -verify jwt-public-new.pem -signature /dev/stdin
   # Should output: Verified OK
   ```

4. **Secure Private Key**

   ```bash
   # Encrypt private key (for storage)
   openssl rsa -in jwt-private-new.pem -out jwt-private-new-encrypted.pem -aes256

   # Set restrictive permissions
   chmod 400 jwt-private-new.pem
   chown temporal:temporal jwt-private-new.pem
   ```

### Phase 2: Add New Public Key to Temporal

1. **Update Temporal Configuration** (Kubernetes)

   ```yaml
   # temporal-jwt-config.yaml
   apiVersion: v1
   kind: ConfigMap
   metadata:
     name: temporal-jwt-keys
     namespace: temporal
   data:
     # OLD KEY (keep during dual-key period)
     jwt-public-old.pem: |
       -----BEGIN PUBLIC KEY-----
       MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA...
       -----END PUBLIC KEY-----

     # NEW KEY (add now)
     jwt-public-new.pem: |
       -----BEGIN PUBLIC KEY-----
       MIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEA...
       -----END PUBLIC KEY-----
   ```

   ```bash
   kubectl apply -f temporal-jwt-config.yaml
   ```

2. **Update Temporal Server Configuration**

   ```yaml
   global:
     authorization:
       authorizer: "default"
       jwtKeyProvider:
         # Support multiple keys during rotation
         keySourceURIs:
           - "file:///etc/temporal/jwt-keys/jwt-public-old.pem"
           - "file:///etc/temporal/jwt-keys/jwt-public-new.pem"  # Add new key
         refreshInterval: "1h"  # Auto-reload keys
   ```

3. **Deploy Configuration Update**

   ```bash
   # Kubernetes
   kubectl apply -f temporal-config.yaml

   # Wait for config reload
   sleep 65  # Wait > refreshInterval

   # Verify keys loaded
   kubectl logs -n temporal deployment/temporal-frontend | \
     grep "JWT.*public key"
   ```

4. **Verify Both Keys Trusted**

   Create test tokens with each key and verify both are accepted:

   ```bash
   # Test with old key
   OLD_TOKEN=$(generate_jwt_token --key jwt-private-old.pem --kid temporal-jwt-20251120)
   temporal workflow list --address temporal.example.com:7233 \
     --tls-ca-path ca.pem \
     --headers "authorization=Bearer $OLD_TOKEN"

   # Test with new key
   NEW_TOKEN=$(generate_jwt_token --key jwt-private-new.pem --kid temporal-jwt-20251122)
   temporal workflow list --address temporal.example.com:7233 \
     --tls-ca-path ca.pem \
     --headers "authorization=Bearer $NEW_TOKEN"

   # Both should succeed
   ```

### Phase 3: Update JWT Signer (Identity Provider)

1. **Deploy New Private Key to IdP**

   ```bash
   # Copy to IdP/auth service
   scp jwt-private-new.pem auth-server:/etc/auth/jwt-keys/

   # Update IdP configuration to use new key
   # (Implementation depends on your IdP - Auth0, Okta, custom, etc.)
   ```

2. **Configure IdP to Use New Key**

   **Example: Custom Auth Service**

   ```javascript
   // auth-service/config.js
   module.exports = {
     jwt: {
       algorithm: 'RS256',
       privateKeyPath: '/etc/auth/jwt-keys/jwt-private-new.pem',
       keyId: 'temporal-jwt-20251122',  // Update kid
       expiresIn: '1h'
     }
   };
   ```

   **Example: Auth0 Dashboard**
   - Go to Applications → API → Settings
   - Upload new signing key
   - Set key rotation schedule

3. **Gradually Roll Out New Key**

   If possible, configure IdP to use both keys with weighted distribution:

   ```yaml
   # Example configuration
   jwt_keys:
     - key_id: temporal-jwt-20251120
       private_key_file: /etc/auth/jwt-keys/jwt-private-old.pem
       weight: 10  # 10% of new tokens

     - key_id: temporal-jwt-20251122
       private_key_file: /etc/auth/jwt-keys/jwt-private-new.pem
       weight: 90  # 90% of new tokens
   ```

   Then gradually shift to 100% new key:
   ```
   Day 0: 90% new, 10% old
   Day 1: 100% new, 0% old
   ```

4. **Verify New Tokens Use New Key**

   ```bash
   # Get new token from IdP
   TOKEN=$(curl -X POST https://auth.example.com/oauth/token \
     -d "grant_type=client_credentials" \
     -d "client_id=..." \
     -d "client_secret=..." | jq -r '.access_token')

   # Decode header to check kid
   echo $TOKEN | cut -d. -f1 | base64 -d | jq .
   # Should show: {"alg":"RS256","kid":"temporal-jwt-20251122","typ":"JWT"}
   ```

### Phase 4: Wait for Old Tokens to Expire

1. **Calculate Wait Period**

   ```bash
   # Find maximum token expiration time
   # Check issued tokens to find the latest expiration

   # Example: If tokens expire in 1 hour
   MAX_TOKEN_TTL="1 hour"

   # Add safety buffer
   WAIT_PERIOD="25 hours"  # 1 hour + 24 hour buffer

   # Set reminder
   at now + 25 hours <<EOF
   echo "JWT key rotation: Ready to remove old key" | \
     mail -s "JWT Key Rotation Phase 4 Complete" ops@example.com
   EOF
   ```

2. **Monitor Token Distribution**

   ```bash
   # Check which keys are being used in logs
   kubectl logs -n temporal deployment/temporal-frontend | \
     grep "JWT token validated" | \
     grep -oP 'kid=\K[^,]+' | \
     sort | uniq -c

   # Output example:
   #  152 temporal-jwt-20251120  (old key - decreasing)
   #  1847 temporal-jwt-20251122  (new key - increasing)
   ```

3. **Wait for Old Key Usage to Reach Zero**

   Monitor until no tokens with old kid are seen for at least 1 hour.

### Phase 5: Remove Old Public Key

1. **Update Temporal Configuration**

   ```yaml
   global:
     authorization:
       jwtKeyProvider:
         keySourceURIs:
           # Remove old key, keep only new
           - "file:///etc/temporal/jwt-keys/jwt-public-new.pem"
         refreshInterval: "1h"
   ```

2. **Remove Old Key from ConfigMap**

   ```yaml
   # temporal-jwt-config.yaml
   apiVersion: v1
   kind: ConfigMap
   metadata:
     name: temporal-jwt-keys
   data:
     # Only new key remains
     jwt-public-new.pem: |
       -----BEGIN PUBLIC KEY-----
       MIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEA...
       -----END PUBLIC KEY-----
   ```

   ```bash
   kubectl apply -f temporal-jwt-config.yaml
   kubectl apply -f temporal-config.yaml
   ```

3. **Verify Cleanup**

   ```bash
   # Old key should be rejected now
   OLD_TOKEN=$(generate_jwt_token --key jwt-private-old.pem --kid temporal-jwt-20251120)
   temporal workflow list --headers "authorization=Bearer $OLD_TOKEN"
   # Should fail with: "permission denied: authentication failed"

   # New key should still work
   NEW_TOKEN=$(generate_jwt_token --key jwt-private-new.pem --kid temporal-jwt-20251122)
   temporal workflow list --headers "authorization=Bearer $NEW_TOKEN"
   # Should succeed
   ```

4. **Archive Old Keys**

   ```bash
   # Move to archive (don't delete immediately)
   mkdir -p /etc/temporal/jwt-keys/archive/$(date +%Y%m%d)
   mv /etc/temporal/jwt-keys/jwt-private-old.pem \
      /etc/temporal/jwt-keys/archive/$(date +%Y%m%d)/

   # Set calendar reminder to delete after 90 days
   ```

---

## Emergency Rotation (Key Compromise)

If a JWT signing key is compromised, follow this expedited procedure:

### Immediate Actions (Within 1 Hour)

1. **Generate New Key Immediately**

   ```bash
   # Generate emergency key
   openssl genrsa -out jwt-private-emergency.pem 4096
   openssl rsa -in jwt-private-emergency.pem -pubout -out jwt-public-emergency.pem
   KID="temporal-jwt-emergency-$(date +%Y%m%d%H%M)"
   ```

2. **Deploy New Public Key to Temporal**

   Skip the dual-key period - replace immediately:

   ```yaml
   jwtKeyProvider:
     keySourceURIs:
       - "file:///etc/temporal/jwt-keys/jwt-public-emergency.pem"
   ```

3. **Force Configuration Reload**

   ```bash
   # Restart Temporal pods to force immediate key reload
   kubectl rollout restart deployment/temporal-frontend
   kubectl rollout restart deployment/temporal-history
   kubectl rollout restart deployment/temporal-matching
   kubectl rollout restart deployment/temporal-worker
   ```

4. **Update IdP Immediately**

   Deploy emergency private key to all token signers.

5. **Revoke Old Key**

   If using key management service, mark old key as revoked:

   ```bash
   # Example: AWS KMS
   aws kms schedule-key-deletion --key-id <old-key-id> --pending-window-in-days 7

   # Example: HashiCorp Vault
   vault delete transit/keys/temporal-jwt-old
   ```

### Post-Incident Actions (Within 24 Hours)

1. **Invalidate All Old Tokens**

   Inform users to re-authenticate:
   ```bash
   # If you track active sessions, invalidate them
   redis-cli KEYS "session:*" | xargs redis-cli DEL
   ```

2. **Audit Key Access**

   ```bash
   # Review who had access to compromised key
   # Check Git history
   git log --all -p -S "jwt-private"

   # Check file access logs
   ausearch -f /etc/auth/jwt-keys/jwt-private-compromised.pem
   ```

3. **Implement Additional Controls**

   - Enable HSM (Hardware Security Module) for key storage
   - Implement key access auditing
   - Add key usage rate limiting

4. **Update Security Documentation**

   Document incident and implement preventive measures.

---

## Automated Rotation

### Using Scheduled Script

```bash
#!/bin/bash
# auto-rotate-jwt-keys.sh - Automated JWT key rotation

set -e

# Configuration
ROTATION_INTERVAL_DAYS=90
KEY_DIR="/etc/temporal/jwt-keys"
CURRENT_KEY="$KEY_DIR/jwt-private-current.pem"
KID_PREFIX="temporal-jwt"

# Generate new key
NEW_KID="${KID_PREFIX}-$(date +%Y%m%d)"
NEW_PRIVATE="$KEY_DIR/jwt-private-new.pem"
NEW_PUBLIC="$KEY_DIR/jwt-public-new.pem"

echo "=== JWT Key Rotation Started ==="
echo "New Key ID: $NEW_KID"

# Generate key pair
openssl genrsa -out "$NEW_PRIVATE" 4096
openssl rsa -in "$NEW_PRIVATE" -pubout -out "$NEW_PUBLIC"
chmod 400 "$NEW_PRIVATE"

# Add to Temporal (dual-key period)
kubectl create configmap temporal-jwt-keys \
  --from-file=jwt-public-old.pem="$KEY_DIR/jwt-public-current.pem" \
  --from-file=jwt-public-new.pem="$NEW_PUBLIC" \
  --dry-run=client -o yaml | kubectl apply -f -

echo "✓ New public key added to Temporal"

# Wait for key propagation
sleep 120

# Update IdP to use new key
update_idp_key "$NEW_PRIVATE" "$NEW_KID"

echo "✓ IdP updated to use new key"

# Calculate wait period (based on max token TTL + buffer)
MAX_TOKEN_TTL_HOURS=1
BUFFER_HOURS=24
WAIT_HOURS=$((MAX_TOKEN_TTL_HOURS + BUFFER_HOURS))

echo "Waiting $WAIT_HOURS hours for old tokens to expire..."

# Schedule old key removal
at now + $WAIT_HOURS hours <<EOF
  # Remove old key from Temporal
  kubectl create configmap temporal-jwt-keys \
    --from-file=jwt-public-new.pem="$NEW_PUBLIC" \
    --dry-run=client -o yaml | kubectl apply -f -

  # Archive old key
  mv "$KEY_DIR/jwt-private-current.pem" \
     "$KEY_DIR/archive/jwt-private-\$(date +%Y%m%d).pem"
  mv "$KEY_DIR/jwt-public-current.pem" \
     "$KEY_DIR/archive/jwt-public-\$(date +%Y%m%d).pem"

  # Promote new key to current
  mv "$NEW_PRIVATE" "$CURRENT_KEY"
  mv "$NEW_PUBLIC" "$KEY_DIR/jwt-public-current.pem"

  echo "✓ JWT key rotation complete"
  notify_team "JWT key rotation completed successfully"
EOF

echo "=== JWT Key Rotation Scheduled ==="
echo "Old key will be removed in $WAIT_HOURS hours"
```

### Scheduled Rotation (Cron)

```cron
# Rotate JWT keys every 90 days at 2 AM
0 2 1 */3 * /usr/local/bin/auto-rotate-jwt-keys.sh >> /var/log/temporal/jwt-rotation.log 2>&1
```

### Using HashiCorp Vault

```hcl
# Configure Vault for automatic JWT key rotation
path "transit/keys/temporal-jwt" {
  capabilities = ["create", "read", "update"]

  # Automatically rotate key every 90 days
  allowed_parameters = {
    auto_rotate_period = ["7776000"]  # 90 days in seconds
  }
}
```

---

## Monitoring and Validation

### Metrics to Monitor

```promql
# Token validation success rate
rate(temporal_auth_token_validated_total{result="success"}[5m]) /
rate(temporal_auth_token_validated_total[5m])

# Track which keys are being used
sum by (kid) (temporal_auth_token_validated_total)

# Authentication failures
rate(temporal_auth_failure_total[5m])
```

### Alerts

```yaml
- alert: JWTKeyRotationNeeded
  expr: (temporal_jwt_key_age_days > 80)
  labels:
    severity: warning
  annotations:
    summary: "JWT signing key rotation needed soon"
    description: "Key {{ $labels.kid }} is {{ $value }} days old"

- alert: JWTValidationFailureSpike
  expr: rate(temporal_auth_failure_total[5m]) > 10
  for: 5m
  labels:
    severity: critical
  annotations:
    summary: "High JWT validation failure rate"
    description: "Check if key rotation caused issues"
```

### Validation Script

```bash
#!/bin/bash
# validate-jwt-keys.sh

TEMPORAL_ENDPOINT="temporal.example.com:7233"
AUTH_ENDPOINT="https://auth.example.com"

echo "=== JWT Key Validation ==="

# Get current key ID from auth service
CURRENT_KID=$(curl -s "$AUTH_ENDPOINT/.well-known/jwks.json" | \
  jq -r '.keys[0].kid')
echo "Current Key ID: $CURRENT_KID"

# Request new token
TOKEN=$(curl -s -X POST "$AUTH_ENDPOINT/oauth/token" \
  -d "grant_type=client_credentials" \
  -d "client_id=$CLIENT_ID" \
  -d "client_secret=$CLIENT_SECRET" | \
  jq -r '.access_token')

# Decode token header
TOKEN_KID=$(echo "$TOKEN" | cut -d. -f1 | base64 -d 2>/dev/null | jq -r '.kid')
echo "Token Key ID: $TOKEN_KID"

# Test token with Temporal
temporal workflow list \
  --address "$TEMPORAL_ENDPOINT" \
  --tls-ca-path /etc/temporal/certs/ca.pem \
  --headers "authorization=Bearer $TOKEN" \
  >/dev/null 2>&1

if [ $? -eq 0 ]; then
  echo "✓ JWT validation: PASSED"
  exit 0
else
  echo "✗ JWT validation: FAILED"
  exit 1
fi
```

---

## Troubleshooting

### Issue: Tokens suddenly rejected after rotation

**Symptoms:**
```
permission denied: authentication failed
```

**Common Causes:**
1. Old key removed before waiting period complete
2. IdP still signing with old key
3. Configuration not reloaded

**Resolution:**
```bash
# Check which keys Temporal currently trusts
kubectl get configmap temporal-jwt-keys -o yaml

# Check which key IdP is using
TOKEN=$(get_new_token)
echo "$TOKEN" | cut -d. -f1 | base64 -d | jq .kid

# If mismatch, re-add old key temporarily
```

### Issue: "Unknown kid in token header"

**Symptoms:**
```
JWT validation failed: unknown key ID
```

**Resolution:**
1. Verify kid in token matches configured keys:
   ```bash
   # Token kid
   echo "$TOKEN" | cut -d. -f1 | base64 -d | jq .kid

   # Configured kids
   kubectl get configmap temporal-jwt-keys -o yaml | grep "kid:"
   ```

2. Ensure Temporal configuration references correct key files
3. Check `refreshInterval` - keys may not be reloaded yet

### Issue: Performance degradation during rotation

**Symptoms:**
- High CPU usage
- Slow token validation

**Likely Cause:**
Multiple public keys causing validation overhead

**Resolution:**
- Keep only 2-3 keys during rotation (old + new + emergency backup)
- Use faster algorithms (ES256 instead of RS256)
- Consider caching validated tokens

---

## Rollback Procedures

If JWT key rotation causes critical issues:

### Immediate Rollback

```bash
# Restore old key as the only trusted key
kubectl create configmap temporal-jwt-keys \
  --from-file=jwt-public-old.pem=/etc/temporal/jwt-keys/archive/jwt-public-old.pem \
  --dry-run=client -o yaml | kubectl apply -f -

# Update Temporal config
kubectl apply -f temporal-config-old.yaml

# Restart to force reload
kubectl rollout restart deployment/temporal-frontend

# Revert IdP to old key
deploy_old_key_to_idp

# Verify
validate-jwt-keys.sh
```

---

## Best Practices

1. **Regular Rotation Schedule**
   - Rotate keys every 90 days minimum
   - More frequently for high-security environments (30-60 days)

2. **Always Use Dual-Key Period**
   - Never skip this unless emergency
   - Wait period must be > max token TTL

3. **Automate Where Possible**
   - Use scheduled scripts or tools like Vault
   - Automate key generation and deployment

4. **Monitor Key Usage**
   - Track which keys are actively used
   - Alert when old keys still in use after rotation

5. **Secure Key Storage**
   - Use HSM or key management service for private keys
   - Never commit keys to version control
   - Encrypt private keys at rest

6. **Test in Staging**
   - Always test rotation procedure in staging
   - Validate scripts and automation

7. **Document Key Lifecycle**
   - Track when each key was generated
   - Document rotation history
   - Maintain key inventory

8. **Emergency Preparedness**
   - Have emergency rotation procedure ready
   - Practice emergency rotation quarterly
   - Keep emergency contact list updated

---

## Key Management Security

### Storing Private Keys

**Best: Hardware Security Module (HSM)**
```bash
# AWS CloudHSM example
aws cloudhsm create-key --cluster-id <cluster-id> --key-spec RSA_4096

# Use PKCS#11 interface
export PKCS11_MODULE_PATH=/opt/cloudhsm/lib/libcloudhsm_pkcs11.so
```

**Good: Encrypted Storage with KMS**
```bash
# Encrypt private key with AWS KMS
aws kms encrypt \
  --key-id alias/temporal-jwt-key \
  --plaintext fileb://jwt-private.pem \
  --output text \
  --query CiphertextBlob | base64 -d > jwt-private-encrypted.bin
```

**Acceptable: File System with Encryption**
```bash
# Encrypt with password
openssl rsa -aes256 -in jwt-private.pem -out jwt-private-encrypted.pem

# Restrictive permissions
chmod 400 jwt-private-encrypted.pem
chown temporal:temporal jwt-private-encrypted.pem
```

### Key Access Auditing

```bash
# Enable file access auditing (Linux)
auditctl -w /etc/temporal/jwt-keys/ -p rwa -k jwt_key_access

# View audit logs
ausearch -k jwt_key_access

# Example: Alert on unexpected access
ausearch -k jwt_key_access | grep -v "temporal" | \
  mail -s "Unexpected JWT key access" security@example.com
```

---

## Related Documentation

- [CERTIFICATE_ROTATION.md](./CERTIFICATE_ROTATION.md) - TLS certificate rotation
- [SECRETS_ROTATION_RUNBOOK.md](./SECRETS_ROTATION_RUNBOOK.md) - Quick reference
- [SECURITY_OPERATOR_GUIDE.md](../../SECURITY_OPERATOR_GUIDE.md) - Authorization configuration

---

**Document Owner:** Security & Infrastructure Team
**Last Reviewed:** 2025-11-22
**Next Review:** 2026-02-22 (quarterly)
