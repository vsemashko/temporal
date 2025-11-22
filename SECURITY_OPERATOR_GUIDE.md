# Temporal Server - Security Configuration Guide for Operators

**Version:** 1.0
**Date:** 2025-11-22
**Target Audience:** DevOps Engineers, SREs, Security Engineers

---

## Table of Contents

1. [Introduction](#introduction)
2. [TLS Configuration](#tls-configuration)
3. [Authorization Configuration](#authorization-configuration)
   - [Authentication Rate Limiting](#authentication-rate-limiting)
4. [Security Best Practices](#security-best-practices)
5. [Migration Guide](#migration-guide)
6. [Troubleshooting](#troubleshooting)
7. [Security Checklist](#security-checklist)

---

## Introduction

This guide provides operators with comprehensive instructions for securely configuring Temporal Server following the recent security enhancements. These changes include:

- **TLS 1.3 as default** with strong cipher suites
- **Enhanced authorization warnings** to prevent misconfiguration
- **HTTP security headers** for web-based protection
- **Improved error handling** to prevent information disclosure

### What's New

The security enhancements introduce new configuration options while maintaining backward compatibility:

- **TLS MinVersion**: Configure minimum TLS version (defaults to 1.3)
- **TLS CipherSuites**: Explicitly control allowed cipher suites
- **Security Headers**: Automatically applied to all HTTP responses
- **Enhanced Warnings**: Clear documentation of security risks

---

## TLS Configuration

### Overview

Temporal Server now defaults to TLS 1.3 with strong cipher suites. This section explains how to configure TLS properly for production environments.

### Basic TLS Configuration

#### Recommended Production Configuration

```yaml
global:
  tls:
    # Refresh interval for certificate rotation
    refreshInterval: 1h

    # Certificate expiration monitoring
    expirationChecks:
      warningWindow: 720h    # 30 days
      errorWindow: 168h      # 7 days
      checkInterval: 12h

    # Frontend TLS (client-facing)
    frontend:
      server:
        certFile: /etc/temporal/certs/frontend-server.pem
        keyFile: /etc/temporal/certs/frontend-server-key.pem
        clientCAFiles:
          - /etc/temporal/certs/ca.pem
        requireClientAuth: true
        # Optional: Override default TLS 1.3 if needed
        # minVersion: "1.3"
        # Optional: Custom cipher suites (use defaults if unsure)
        # cipherSuites:
        #   - "TLS_AES_256_GCM_SHA384"
        #   - "TLS_CHACHA20_POLY1305_SHA256"

      client:
        serverName: temporal.example.com
        rootCaFiles:
          - /etc/temporal/certs/ca.pem
        # CRITICAL: Enable host verification in production
        disableHostVerification: false

    # Internode TLS (service-to-service)
    internode:
      server:
        certFile: /etc/temporal/certs/internode-server.pem
        keyFile: /etc/temporal/certs/internode-server-key.pem
        clientCAFiles:
          - /etc/temporal/certs/ca.pem
        requireClientAuth: true

      client:
        serverName: temporal-internal.example.com
        rootCaFiles:
          - /etc/temporal/certs/ca.pem
        disableHostVerification: false
```

### TLS Version Configuration

#### Default Behavior (Recommended)

By default, Temporal now uses TLS 1.3 as the minimum version. This is the recommended setting for maximum security.

**No configuration needed** - TLS 1.3 is automatic.

#### Backward Compatibility (TLS 1.2)

If you have legacy clients that don't support TLS 1.3, you can configure TLS 1.2:

```yaml
global:
  tls:
    frontend:
      server:
        certFile: /etc/temporal/certs/frontend.pem
        keyFile: /etc/temporal/certs/frontend-key.pem
        minVersion: "1.2"  # Allow TLS 1.2 for backward compatibility
        # Strong cipher suites are still enforced
```

**⚠️ WARNING:** Only use TLS 1.2 if absolutely necessary for client compatibility. Plan to upgrade clients to TLS 1.3.

### Cipher Suite Configuration

#### Using Defaults (Recommended)

The default cipher suites are secure and tested. **No configuration needed.**

Default cipher suites for TLS 1.2 backward compatibility:
- `TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256`
- `TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384`
- `TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256`
- `TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384`
- `TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256`
- `TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256`

TLS 1.3 cipher suites (not configurable, always secure):
- `TLS_AES_128_GCM_SHA256`
- `TLS_AES_256_GCM_SHA384`
- `TLS_CHACHA20_POLY1305_SHA256`

#### Custom Cipher Suites (Advanced)

For compliance or specific security requirements:

```yaml
global:
  tls:
    frontend:
      server:
        certFile: /etc/temporal/certs/frontend.pem
        keyFile: /etc/temporal/certs/frontend-key.pem
        minVersion: "1.2"
        cipherSuites:
          - "TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384"
          - "TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384"
```

**⚠️ CAUTION:** Only configure custom cipher suites if you have specific compliance requirements. Incorrect configuration can weaken security or break client connections.

### Host Verification

#### Production (REQUIRED)

**ALWAYS enable host verification in production:**

```yaml
global:
  tls:
    frontend:
      client:
        serverName: temporal.example.com  # Must match certificate SAN
        disableHostVerification: false    # DEFAULT - explicitly set for clarity
```

#### Development/Testing Only

**ONLY for development environments:**

```yaml
global:
  tls:
    frontend:
      client:
        serverName: temporal-dev.local
        disableHostVerification: true  # ⚠️ DEVELOPMENT ONLY - NEVER IN PRODUCTION
```

**🔴 CRITICAL WARNING:**

Disabling host verification exposes your deployment to man-in-the-middle (MITM) attacks where an attacker can intercept and modify traffic between services. This should **NEVER** be done in production environments.

### Database TLS Configuration

Temporal supports TLS for database connections:

```yaml
persistence:
  defaultStore: default
  datastores:
    default:
      sql:
        pluginName: "postgres"
        databaseName: "temporal"
        connectAddr: "postgres.example.com:5432"
        connectAttributes:
          sslmode: "verify-full"
        tls:
          enabled: true
          enableHostVerification: true  # ⚠️ Set to true in production
          serverName: "postgres.example.com"
          caFile: /etc/temporal/certs/postgres-ca.pem
          # Optional: Client certificate for mTLS
          certFile: /etc/temporal/certs/postgres-client.pem
          keyFile: /etc/temporal/certs/postgres-client-key.pem
```

---

## Authorization Configuration

### Overview

Temporal supports JWT-based authorization with role-based access control (RBAC). Proper authorization configuration is critical for production security.

### Production Authorization (REQUIRED)

**NEVER use noop authorizer in production:**

```yaml
global:
  authorization:
    authorizer: "default"  # Use the default JWT-based authorizer
    permissionsClaimName: "permissions"
    jwtKeyProvider:
      keySourceURIs:
        - "file:///etc/temporal/jwt-public-key.pem"
      refreshInterval: "1h"
```

### JWT Token Format

Temporal expects JWT tokens with the following claims:

```json
{
  "sub": "user@example.com",
  "aud": "temporal",
  "exp": 1735689600,
  "iat": 1735603200,
  "permissions": [
    "system:admin",
    "namespace1:write",
    "namespace2:read"
  ]
}
```

**Permission Format:** `<namespace>:<role>`

**Roles:**
- `read` - Read-only operations
- `write` - Write operations (start workflows, signal, etc.)
- `worker` - Worker operations (poll for tasks)
- `admin` - Administrative operations (create namespaces, etc.)

**System-Level Permissions:**
- `system:read` - Read across all namespaces
- `system:write` - Write across all namespaces
- `system:admin` - Admin across all namespaces

### Noop Authorizer (DEVELOPMENT ONLY)

**🔴 CRITICAL WARNING:**

The noop authorizer bypasses ALL authorization checks. Using this in production is a **CRITICAL SECURITY VULNERABILITY**.

```yaml
# ⚠️ DEVELOPMENT ONLY - NEVER USE IN PRODUCTION
global:
  authorization:
    authorizer: ""  # Empty string = noop authorizer
```

**When you see this configuration:**
- 🔴 All API operations are accessible to anyone
- 🔴 No authentication required
- 🔴 Admin operations exposed
- 🔴 Data can be read/modified/deleted by anyone

**Use only for:**
- Local development
- Internal testing
- Proof-of-concept demos

**NEVER use for:**
- Production deployments
- Staging environments with real data
- Any environment accessible from the internet
- Any environment with compliance requirements

### Authentication Rate Limiting

**NEW: Phase 2 Enhancement**

Authentication rate limiting protects against brute force attacks by tracking failed authentication attempts and temporarily locking out IP addresses that exceed the failure threshold.

#### How It Works

1. **Failure Tracking**: Each failed authentication attempt is tracked by client IP address
2. **Sliding Window**: Failures are counted within a 1-minute sliding window
3. **Lockout**: When an IP exceeds the threshold, it's locked out for a configured duration
4. **Success Reset**: Successful authentication clears all tracked failures for that IP
5. **Automatic Cleanup**: Old tracking data is automatically removed to prevent memory growth

#### Configuration

```yaml
global:
  authorization:
    rateLimit:
      # Enable authentication rate limiting (recommended for production)
      enabled: true

      # Maximum failed authentication attempts per minute before lockout
      # Default: 10
      # Recommended: 5-10 for production, lower for high-security environments
      maxFailuresPerMinute: 10

      # Duration to lock out an IP after exceeding the threshold
      # Default: 5m
      # Recommended: 5m-15m (longer for high-security environments)
      lockoutDuration: 5m
```

#### Example Configurations

**High Security Environment:**
```yaml
global:
  authorization:
    rateLimit:
      enabled: true
      maxFailuresPerMinute: 3
      lockoutDuration: 15m
```

**Standard Production:**
```yaml
global:
  authorization:
    rateLimit:
      enabled: true
      maxFailuresPerMinute: 10
      lockoutDuration: 5m
```

**Development (Disabled):**
```yaml
global:
  authorization:
    rateLimit:
      enabled: false
```

#### Monitoring

Monitor these metrics to track authentication security:

**Metrics:**
- `auth_failure_total` - Total authentication failures
- `auth_rate_limited_total` - Requests blocked by rate limiting
- `auth_lockout_total` - IP addresses locked out
- `auth_tracked_ips` - Current number of IPs being tracked
- `auth_tracker_overflow` - Attempts to track when limit reached

**Example Prometheus Queries:**
```promql
# Authentication failure rate
rate(auth_failure_total[5m])

# Lockout events
increase(auth_lockout_total[1h])

# Currently tracked IPs
auth_tracked_ips

# Rate limiting effectiveness
rate(auth_rate_limited_total[5m]) / rate(auth_failure_total[5m])
```

#### Alerting Recommendations

Set up alerts for security events:

```yaml
# High authentication failure rate
- alert: HighAuthFailureRate
  expr: rate(auth_failure_total[5m]) > 10
  for: 5m
  annotations:
    summary: "High authentication failure rate detected"

# Many IPs being locked out (possible attack)
- alert: MassiveBruteForceAttempt
  expr: rate(auth_lockout_total[5m]) > 5
  for: 5m
  annotations:
    summary: "Multiple IPs being locked out - possible distributed attack"

# Tracker overflow (may need to increase limit)
- alert: AuthTrackerOverflow
  expr: rate(auth_tracker_overflow[5m]) > 1
  for: 5m
  annotations:
    summary: "Auth rate limiter tracker overflow"
```

#### Operational Considerations

**IP Extraction:**
- Rate limiting uses the client IP address from gRPC peer context
- Works correctly with direct connections
- With load balancers/proxies, ensure `X-Forwarded-For` headers are properly configured
- If IP cannot be extracted, requests are allowed (fail-open behavior)

**Memory Usage:**
- Each tracked IP uses approximately 64 bytes
- Default limit: 10,000 IPs tracked = ~640 KB memory
- Old tracking data is automatically cleaned up every 10 minutes
- IPs are removed when: lockout expires AND no failures for 10+ minutes

**Performance Impact:**
- Minimal overhead: RWMutex for thread safety
- Read lock for checking lockouts (fast path)
- Write lock only for recording failures (slow path)
- No database queries - all in-memory

**False Positives:**
- NAT/CGNAT scenarios: Multiple users behind same IP may trigger lockouts
- Corporate networks: Large offices sharing one external IP
- Consider higher thresholds or disabling for known internal IPs
- Monitor `auth_lockout_total` for patterns

#### Troubleshooting

**Users reporting authentication failures:**

1. Check if IP is locked out:
```bash
# Search logs for the user's IP
grep "IP locked out" temporal.log | grep "192.168.1.100"
```

2. Check metrics for that IP's failure count:
```promql
auth_tracked_ips{ip="192.168.1.100"}
```

3. Temporary mitigation:
   - Restart Temporal server (clears all tracking data)
   - Or wait for lockout duration to expire
   - Or increase `maxFailuresPerMinute` temporarily

**Legitimate traffic being blocked:**

Possible causes:
- NAT/CGNAT: Multiple users sharing IP
- Misconfigured clients: Retrying with wrong credentials
- Password sync issues: Old credentials being used

Solutions:
- Increase `maxFailuresPerMinute` threshold
- Increase `lockoutDuration` (counter-intuitive but reduces frequency of lockouts)
- Fix underlying authentication issues
- Consider IP whitelisting for known internal ranges (future enhancement)

**Rate limiter not working:**

Check configuration:
```yaml
# Ensure enabled is true
global:
  authorization:
    rateLimit:
      enabled: true  # Must be explicitly true
```

Check logs for warnings:
```bash
grep "auth rate limit" temporal.log
```

Verify interceptor is loaded:
```bash
# Should see "AuthRateLimitInterceptor" in startup logs
grep "AuthRateLimitInterceptor" temporal.log
```

---

## Security Best Practices

### Certificate Management

#### Certificate Generation

Use proper certificate management:

```bash
# Generate CA (one-time)
openssl req -new -x509 -days 3650 -keyout ca-key.pem -out ca.pem \
  -subj "/CN=Temporal CA"

# Generate server certificate
openssl genrsa -out server-key.pem 2048
openssl req -new -key server-key.pem -out server.csr \
  -subj "/CN=temporal.example.com"

# Sign with CA
openssl x509 -req -in server.csr -CA ca.pem -CAkey ca-key.pem \
  -CAcreateserial -out server.pem -days 365 \
  -extensions SAN -extfile <(printf "\n[SAN]\nsubjectAltName=DNS:temporal.example.com,DNS:*.temporal.example.com")
```

#### File Permissions

Secure your certificate files:

```bash
# CA certificate (public)
chmod 444 /etc/temporal/certs/ca.pem

# Server certificates (public)
chmod 444 /etc/temporal/certs/*.pem

# Private keys (secret)
chmod 400 /etc/temporal/certs/*-key.pem
chown temporal:temporal /etc/temporal/certs/*-key.pem
```

#### Certificate Rotation

Temporal supports zero-downtime certificate rotation:

1. Place new certificates in the certificate directory
2. Temporal automatically reloads based on `refreshInterval`
3. Old connections continue using old certs
4. New connections use new certs
5. Graceful migration with no downtime

Configure monitoring for certificate expiration:

```yaml
global:
  tls:
    expirationChecks:
      warningWindow: 720h   # 30 days - emit warnings
      errorWindow: 168h     # 7 days - emit errors
      checkInterval: 12h    # Check every 12 hours
```

### Secrets Management

#### DO NOT store secrets in configuration files

**❌ BAD - Secrets in config:**
```yaml
persistence:
  defaultStore: default
  datastores:
    default:
      sql:
        connectAddr: "postgres.example.com:5432"
        connectAttributes:
          user: "temporal"
          password: "hardcoded-password"  # ❌ NEVER DO THIS
```

**✅ GOOD - Use environment variables:**
```yaml
persistence:
  defaultStore: default
  datastores:
    default:
      sql:
        connectAddr: "postgres.example.com:5432"
        connectAttributes:
          user: "${DB_USER}"
          password: "${DB_PASSWORD}"
```

```bash
# Inject secrets from secret management system
export DB_USER="temporal"
export DB_PASSWORD="$(vault read -field=password secret/temporal/db)"
```

#### Recommended Secrets Management Solutions

- **HashiCorp Vault** - Enterprise secrets management
- **AWS Secrets Manager** - AWS native solution
- **Azure Key Vault** - Azure native solution
- **Google Secret Manager** - GCP native solution
- **Kubernetes Secrets** - For K8s deployments (with encryption at rest)

### Network Security

#### Firewall Rules

Restrict access to Temporal services:

```bash
# Frontend (gRPC) - port 7233
# Allow from: Client applications, Workers
iptables -A INPUT -p tcp --dport 7233 -s 10.0.0.0/8 -j ACCEPT
iptables -A INPUT -p tcp --dport 7233 -j DROP

# Frontend (HTTP) - port 7243 (if enabled)
# Allow from: Load balancer, monitoring
iptables -A INPUT -p tcp --dport 7243 -s 10.0.0.0/8 -j ACCEPT
iptables -A INPUT -p tcp --dport 7243 -j DROP

# Internode - port 7234
# Allow from: Only other Temporal services
iptables -A INPUT -p tcp --dport 7234 -s 10.0.1.0/24 -j ACCEPT
iptables -A INPUT -p tcp --dport 7234 -j DROP
```

#### Load Balancer Configuration

If using a load balancer, enable:
- **TLS termination** at the load balancer
- **Client certificate validation** (for mTLS)
- **Request size limits** (prevent DoS)
- **Rate limiting** (prevent abuse)

### Monitoring and Alerting

Set up alerts for security events:

```yaml
# Metrics to monitor
- temporal_authorization_failed_count     # Failed auth attempts
- temporal_tls_handshake_failures         # TLS issues
- temporal_certificate_expiration_days    # Cert expiration
- temporal_authentication_errors          # Auth errors
```

**Recommended Alerts:**
1. Certificate expiring in < 30 days
2. Spike in authentication failures (> 100/min)
3. TLS handshake failure rate > 1%
4. Noop authorizer active (should never happen in production)

---

## Migration Guide

### Upgrading to TLS 1.3

#### Step 1: Verify Client Compatibility

Before upgrading, verify all clients support TLS 1.3:

**Go SDK:** v1.12+ supports TLS 1.3
**Java SDK:** v1.13+ supports TLS 1.3
**Python SDK:** v1.2+ supports TLS 1.3
**TypeScript SDK:** v1.5+ supports TLS 1.3

#### Step 2: Test in Staging

1. Deploy updated Temporal server to staging
2. Run full integration test suite
3. Verify TLS handshakes succeed
4. Check for connection errors in logs
5. Monitor TLS version metrics

```bash
# Verify TLS 1.3 is negotiated
openssl s_client -connect temporal.example.com:7233 -tls1_3
```

#### Step 3: Gradual Production Rollout

1. **Week 1:** Deploy to 10% of production (canary)
   - Monitor error rates
   - Check client connection success

2. **Week 2:** Deploy to 50% of production
   - Continue monitoring
   - Verify no TLS 1.2 clients failing

3. **Week 3:** Deploy to 100% of production
   - Full rollout
   - 24/7 monitoring for first 48 hours

#### Step 4: Rollback Plan

If issues arise:

```yaml
# Emergency rollback to TLS 1.2
global:
  tls:
    frontend:
      server:
        minVersion: "1.2"  # Temporary rollback
```

Redeploy and investigate client compatibility issues.

### Migrating from Noop to Proper Authorization

If currently using noop authorizer, follow this migration path:

#### Step 1: Implement JWT Token Generation

Set up a JWT token service:

```go
import "github.com/golang-jwt/jwt/v4"

func generateToken(userID string, permissions []string) (string, error) {
    token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
        "sub": userID,
        "aud": "temporal",
        "exp": time.Now().Add(time.Hour * 24).Unix(),
        "iat": time.Now().Unix(),
        "permissions": permissions,
    })

    return token.SignedString(privateKey)
}
```

#### Step 2: Update Temporal Configuration

```yaml
global:
  authorization:
    authorizer: "default"
    permissionsClaimName: "permissions"
    jwtKeyProvider:
      keySourceURIs:
        - "file:///etc/temporal/jwt-public-key.pem"
```

#### Step 3: Update All Clients

Ensure all clients include JWT tokens:

```go
// Go SDK example
client, err := client.NewClient(client.Options{
    HostPort: "temporal.example.com:7233",
    Credentials: client.NewAPIKeyDynamicAuthProvider(func() (string, error) {
        return getJWTToken(), nil
    }),
})
```

#### Step 4: Enable Authorization

1. Deploy configuration with authorization enabled
2. Verify clients can authenticate
3. Test authorization denials work correctly
4. Monitor auth failure rates

---

## Troubleshooting

### TLS Handshake Failures

**Symptom:** Clients cannot connect, TLS handshake errors

**Diagnosis:**
```bash
# Test TLS connection
openssl s_client -connect temporal.example.com:7233 -showcerts

# Check supported TLS versions
openssl s_client -connect temporal.example.com:7233 -tls1_2
openssl s_client -connect temporal.example.com:7233 -tls1_3
```

**Common Causes:**
1. Client doesn't support TLS 1.3
   - **Solution:** Configure `minVersion: "1.2"` or upgrade client
2. Certificate hostname mismatch
   - **Solution:** Ensure certificate SAN matches `serverName`
3. Expired certificate
   - **Solution:** Rotate certificates
4. Wrong CA certificate
   - **Solution:** Verify `rootCaFiles` points to correct CA

### Authentication Failures

**Symptom:** "authentication failed" errors

**Diagnosis:**
```bash
# Check server logs for detailed error
grep "JWT parsing failed" /var/log/temporal/server.log
grep "invalid authorization" /var/log/temporal/server.log
```

**Common Causes:**
1. Missing or malformed JWT token
   - **Solution:** Verify token format, ensure "Bearer " prefix
2. Expired token
   - **Solution:** Check token expiration, implement refresh logic
3. Wrong signing key
   - **Solution:** Verify public key matches private key used to sign
4. Missing required claims
   - **Solution:** Ensure token has "sub" and "permissions" claims

### Certificate Verification Failures

**Symptom:** "certificate verify failed" errors

**Diagnosis:**
```bash
# Verify certificate chain
openssl verify -CAfile ca.pem server.pem

# Check certificate details
openssl x509 -in server.pem -text -noout
```

**Common Causes:**
1. Certificate not trusted by CA
   - **Solution:** Ensure certificate is signed by configured CA
2. Hostname mismatch
   - **Solution:** Verify SAN/CN matches `serverName`
3. Certificate expired
   - **Solution:** Check expiration date, rotate if needed

### HTTP Security Headers Breaking Clients

**Symptom:** Web UI or API clients failing after upgrade

**Diagnosis:**
Check browser console for CSP violations

**Solution:**
The security headers are designed to be restrictive. If legitimate clients are blocked:

1. Review the specific header causing issues
2. Consider if the client should be updated to comply
3. As a last resort, headers can be customized (requires code change)

---

## Security Checklist

### Pre-Production Deployment

- [ ] TLS 1.3 enabled (or TLS 1.2 with justification)
- [ ] Strong cipher suites configured (or using defaults)
- [ ] Host verification enabled (`disableHostVerification: false`)
- [ ] Valid TLS certificates with correct SANs
- [ ] Certificate expiration monitoring configured
- [ ] Proper authorization configured (`authorizer: "default"`)
- [ ] JWT token provider implemented and tested
- [ ] Secrets managed externally (not in config files)
- [ ] Database connections use TLS
- [ ] Firewall rules restrict access
- [ ] Monitoring and alerting configured
- [ ] Security headers tested with clients
- [ ] All test credentials removed from production config

### Periodic Security Review (Monthly)

- [ ] Review certificate expiration dates
- [ ] Audit authorization failures
- [ ] Review TLS handshake failure rate
- [ ] Check for outdated cipher suites
- [ ] Verify no noop authorizer in production
- [ ] Review access logs for anomalies
- [ ] Update dependencies for security patches
- [ ] Test certificate rotation procedure
- [ ] Review firewall rules
- [ ] Audit user permissions

### Incident Response

If a security incident is detected:

1. **Isolate:** Immediately isolate affected systems
2. **Assess:** Determine scope and impact
3. **Contain:** Revoke compromised credentials, rotate certificates
4. **Eradicate:** Remove malicious access, patch vulnerabilities
5. **Recover:** Restore from clean backups if needed
6. **Review:** Conduct post-incident review

**Emergency Contacts:**
- Security team: security@example.com
- On-call engineer: oncall@example.com
- Incident commander: incidents@example.com

---

## Additional Resources

- **Official Documentation:** https://docs.temporal.io/security
- **Security Advisories:** https://github.com/temporalio/temporal/security/advisories
- **TLS Best Practices:** https://wiki.mozilla.org/Security/Server_Side_TLS
- **OWASP Guidelines:** https://cheatsheetseries.owasp.org/

---

**Document Version:** 1.0
**Last Updated:** 2025-11-22
**Maintained By:** Security Team
**Questions:** security@temporal.io
